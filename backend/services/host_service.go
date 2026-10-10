package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

type HostService struct{}

func NewHostService() *HostService {
	return &HostService{}
}

// validatePersonID 校验人员关联，personID 为 nil 表示解除关联
func validatePersonID(personID *uint) error {
	if personID == nil {
		return nil
	}
	var count int64
	database.DB.Model(&models.Person{}).Where("id = ?", *personID).Count(&count)
	if count == 0 {
		return ErrPersonNotFound
	}
	return nil
}

type ListHostResponse struct {
	Hosts []models.Host `json:"hosts"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"page_size"`
}

type FilterHostRequest struct {
	Page           int    `form:"page,default=1"`
	PageSize       int    `form:"page_size,default=20"`
	Keyword        string `form:"keyword"`
	EnvType        string `form:"env_type"`
	AssetType      string `form:"asset_type"`
	CPUArch        string `form:"cpu_arch"`
	IsDBServer     string `form:"is_db_server"`
	Status         string `form:"status"`
	Region         string `form:"region"`
	ApplicantEmpty string `form:"applicant_empty"`
}

func (s *HostService) Filter(req FilterHostRequest) (*ListHostResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	db := database.DB.Model(&models.Host{})

	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		subQuery := database.DB.Model(&models.HostApplication{}).
			Select("host_id").
			Where(`(apply_unit ILIKE ? OR applicant ILIKE ? OR project ILIKE ? OR remark ILIKE ?)`,
				like, like, like, like)
		db = db.Where(`
			(region ILIKE ? OR instance_id ILIKE ? OR name ILIKE ? OR private_ip ILIKE ? OR
			 os ILIKE ? OR status ILIKE ? OR tags ILIKE ?
			 OR hosts.id IN (?))
		`, like, like, like, like, like, like, like, subQuery)
	}

	if req.EnvType != "" {
		db = db.Where("env_type = ?", req.EnvType)
	}
	if req.AssetType != "" {
		db = db.Where("asset_type = ?", req.AssetType)
	}
	if req.CPUArch != "" {
		db = db.Where("cpu_arch = ?", req.CPUArch)
	}
	if req.IsDBServer != "" {
		if req.IsDBServer == "1" {
			db = db.Where("is_db_server = ?", true)
		} else {
			db = db.Where("is_db_server = ?", false)
		}
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Region != "" {
		db = db.Where("region = ?", req.Region)
	}
	if req.ApplicantEmpty != "" {
		switch req.ApplicantEmpty {
		case "1":
			db = db.Where(`
				hosts.id IN (SELECT ha.host_id FROM host_applications ha WHERE COALESCE(ha.applicant, '') = '')
				OR hosts.id NOT IN (SELECT ha2.host_id FROM host_applications ha2)
			`)
		case "0":
			db = db.Where(`
				hosts.id IN (SELECT ha.host_id FROM host_applications ha WHERE COALESCE(ha.applicant, '') != '')
			`)
		}
	}

	var total int64
	db.Count(&total)

	var hosts []models.Host
	offset := (req.Page - 1) * req.PageSize
	result := db.Preload("Application").Preload("Person").Offset(offset).Limit(req.PageSize).Order("id ASC").Find(&hosts)
	if result.Error != nil {
		return nil, result.Error
	}

	return &ListHostResponse{
		Hosts: hosts,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

func (s *HostService) ListRegions() ([]string, error) {
	var regions []string
	result := database.DB.Model(&models.Host{}).Distinct().Where("region != ''").Pluck("region", &regions)
	if result.Error != nil {
		return nil, result.Error
	}
	return regions, nil
}

func (s *HostService) GetByID(id uint) (*models.Host, error) {
	var host models.Host
	result := database.DB.
		Preload("Application").
		Preload("Person").
		Preload("PortMappings").
		First(&host, id)
	if result.Error != nil {
		return nil, result.Error
	}
	zeroTrusts, err := ZeroTrustsByHost(id)
	if err != nil {
		return nil, err
	}
	host.ZeroTrusts = zeroTrusts
	return &host, nil
}

// OptionalUint 区分「未传」与「显式传 null」，用于可清空的关联字段
type OptionalUint struct {
	Value   *uint
	Present bool
}

func (o *OptionalUint) UnmarshalJSON(b []byte) error {
	o.Present = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var v uint
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}

type CreateHostRequest struct {
	Region     string       `json:"region" binding:"required"`
	InstanceID string       `json:"instance_id"`
	Name       string       `json:"name" binding:"required"`
	PrivateIP  string       `json:"private_ip" binding:"required"`
	AssetType  string       `json:"asset_type"`
	OS         string       `json:"os"`
	CPU        int          `json:"cpu"`
	CPUArch    string       `json:"cpu_arch"`
	Memory     int          `json:"memory"`
	Disk       int          `json:"disk"`
	SystemDisk int          `json:"system_disk"`
	DataDisk   int          `json:"data_disk"`
	EnvType    string       `json:"env_type"`
	IsDBServer *bool        `json:"is_db_server"`
	Status     string       `json:"status"`
	OpenPorts  string       `json:"open_ports"`
	Tags       string       `json:"tags"`
	PersonID   OptionalUint `json:"person_id"`

	ApplyUnit         string `json:"apply_unit"`
	Applicant         string `json:"applicant"`
	ApplicantContact  string `json:"applicant_contact"`
	Project           string `json:"project"`
	ApplyReason       string `json:"apply_reason"`
	ApplyConfig       string `json:"apply_config"`
	ApplyTime         string `json:"apply_time"`
	ObjectStorageSize string `json:"object_storage_size"`
	Remark            string `json:"remark"`
}

func (s *HostService) Create(req CreateHostRequest) (uint, error) {
	req.ApplyTime = strings.TrimSpace(req.ApplyTime)
	if err := utils.ValidateDateOnly("申请时间", req.ApplyTime); err != nil {
		return 0, err
	}

	// Check unique private_ip
	var count int64
	database.DB.Model(&models.Host{}).Where("private_ip = ?", req.PrivateIP).Count(&count)
	if count > 0 {
		return 0, errors.New("内网IP已存在")
	}

	tx := database.DB.Begin()

	host := models.Host{
		Region:     req.Region,
		InstanceID: req.InstanceID,
		Name:       req.Name,
		PrivateIP:  req.PrivateIP,
		AssetType:  req.AssetType,
		OS:         req.OS,
		CPU:        req.CPU,
		CPUArch:    req.CPUArch,
		Memory:     req.Memory,
		Disk:       req.Disk,
		SystemDisk: req.SystemDisk,
		DataDisk:   req.DataDisk,
		EnvType:    req.EnvType,
		Status:     req.Status,
		OpenPorts:  req.OpenPorts,
		Tags:       req.Tags,
	}

	if req.IsDBServer != nil {
		host.IsDBServer = *req.IsDBServer
	}

	if req.PersonID.Present {
		if err := validatePersonID(req.PersonID.Value); err != nil {
			return 0, err
		}
		host.PersonID = req.PersonID.Value
	}

	if err := tx.Create(&host).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	app := models.HostApplication{
		HostID:            host.ID,
		ApplyUnit:         req.ApplyUnit,
		Applicant:         req.Applicant,
		ApplicantContact:  req.ApplicantContact,
		Project:           req.Project,
		ApplyReason:       req.ApplyReason,
		ApplyConfig:       req.ApplyConfig,
		ApplyTime:         req.ApplyTime,
		ObjectStorageSize: req.ObjectStorageSize,
		Remark:            req.Remark,
	}

	if err := tx.Create(&app).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()
	return host.ID, nil
}

type UpdateHostRequest struct {
	Region     string       `json:"region"`
	InstanceID string       `json:"instance_id"`
	Name       string       `json:"name"`
	PrivateIP  string       `json:"private_ip"`
	AssetType  string       `json:"asset_type"`
	OS         string       `json:"os"`
	CPU        int          `json:"cpu"`
	CPUArch    string       `json:"cpu_arch"`
	Memory     int          `json:"memory"`
	Disk       int          `json:"disk"`
	SystemDisk int          `json:"system_disk"`
	DataDisk   int          `json:"data_disk"`
	EnvType    string       `json:"env_type"`
	IsDBServer *bool        `json:"is_db_server"`
	Status     string       `json:"status"`
	OpenPorts  string       `json:"open_ports"`
	Tags       string       `json:"tags"`
	PersonID   OptionalUint `json:"person_id"`

	ApplyUnit         *string `json:"apply_unit"`
	Applicant         *string `json:"applicant"`
	ApplicantContact  *string `json:"applicant_contact"`
	Project           *string `json:"project"`
	ApplyReason       *string `json:"apply_reason"`
	ApplyConfig       *string `json:"apply_config"`
	ApplyTime         *string `json:"apply_time"`
	ObjectStorageSize *string `json:"object_storage_size"`
	Remark            *string `json:"remark"`
}

func (s *HostService) Update(id uint, req UpdateHostRequest) error {
	var host models.Host
	if err := database.DB.First(&host, id).Error; err != nil {
		return err
	}

	// Check unique private_ip if changed
	if req.PrivateIP != "" && req.PrivateIP != host.PrivateIP {
		var count int64
		database.DB.Model(&models.Host{}).Where("private_ip = ? AND id != ?", req.PrivateIP, id).Count(&count)
		if count > 0 {
			return errors.New("内网IP已存在")
		}
	}

	if req.ApplyTime != nil {
		v := strings.TrimSpace(*req.ApplyTime)
		if err := utils.ValidateDateOnly("申请时间", v); err != nil {
			return err
		}
		*req.ApplyTime = v
	}

	tx := database.DB.Begin()

	updates := map[string]interface{}{}
	if req.Region != "" {
		updates["region"] = req.Region
	}
	if req.InstanceID != "" {
		updates["instance_id"] = req.InstanceID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.PrivateIP != "" {
		updates["private_ip"] = req.PrivateIP
	}
	if req.AssetType != "" {
		updates["asset_type"] = req.AssetType
	}
	if req.OS != "" {
		updates["os"] = req.OS
	}
	if req.CPU > 0 {
		updates["cpu"] = req.CPU
	}
	if req.CPUArch != "" {
		updates["cpu_arch"] = req.CPUArch
	}
	if req.Memory > 0 {
		updates["memory"] = req.Memory
	}
	if req.Disk > 0 {
		updates["disk"] = req.Disk
	}
	if req.SystemDisk > 0 {
		updates["system_disk"] = req.SystemDisk
	}
	if req.DataDisk > 0 {
		updates["data_disk"] = req.DataDisk
	}
	if req.EnvType != "" {
		updates["env_type"] = req.EnvType
	}
	if req.IsDBServer != nil {
		updates["is_db_server"] = *req.IsDBServer
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.OpenPorts != "" {
		updates["open_ports"] = req.OpenPorts
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.PersonID.Present {
		if err := validatePersonID(req.PersonID.Value); err != nil {
			return err
		}
		updates["person_id"] = req.PersonID.Value
	}

	if len(updates) > 0 {
		if err := tx.Model(&host).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update application fields
	var app models.HostApplication
	if err := tx.Where("host_id = ?", id).First(&app).Error; err == nil {
		appUpdates := map[string]interface{}{}
		if req.ApplyUnit != nil {
			appUpdates["apply_unit"] = *req.ApplyUnit
		}
		if req.Applicant != nil {
			appUpdates["applicant"] = *req.Applicant
		}
		if req.ApplicantContact != nil {
			appUpdates["applicant_contact"] = *req.ApplicantContact
		}
		if req.Project != nil {
			appUpdates["project"] = *req.Project
		}
		if req.ApplyReason != nil {
			appUpdates["apply_reason"] = *req.ApplyReason
		}
		if req.ApplyConfig != nil {
			appUpdates["apply_config"] = *req.ApplyConfig
		}
		if req.ApplyTime != nil {
			appUpdates["apply_time"] = *req.ApplyTime
		}
		if req.ObjectStorageSize != nil {
			appUpdates["object_storage_size"] = *req.ObjectStorageSize
		}
		if req.Remark != nil {
			appUpdates["remark"] = *req.Remark
		}
		if len(appUpdates) > 0 {
			tx.Model(&app).Updates(appUpdates)
		}
	}

	tx.Commit()
	return nil
}

func (s *HostService) Delete(id uint) error {
	var host models.Host
	if err := database.DB.First(&host, id).Error; err != nil {
		return err
	}

	refCount, err := HostReferencedByZeroTrust(id)
	if err != nil {
		return err
	}
	if refCount > 0 {
		return fmt.Errorf("%w，%d 条零信任台账记录正在使用", ErrHostReferenced, refCount)
	}

	mappingRef, err := HostReferencedByMapping(id)
	if err != nil {
		return err
	}
	if mappingRef > 0 {
		return fmt.Errorf("%w，%d 条映射台账记录正在使用", ErrHostReferencedByMapping, mappingRef)
	}

	tx := database.DB.Begin()
	tx.Where("host_id = ?", id).Delete(&models.HostApplication{})
	tx.Delete(&host)
	tx.Commit()
	return nil
}
