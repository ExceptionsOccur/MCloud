package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

type HostService struct{}

func NewHostService() *HostService {
	return &HostService{}
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
			 public_ip ILIKE ? OR os ILIKE ? OR status ILIKE ? OR tags ILIKE ?
			 OR hosts.id IN (?))
		`, like, like, like, like, like, like, like, like, subQuery)
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
		if req.ApplicantEmpty == "1" {
			db = db.Where(`
				hosts.id IN (SELECT ha.host_id FROM host_applications ha WHERE COALESCE(ha.applicant, '') = '')
				OR hosts.id NOT IN (SELECT ha2.host_id FROM host_applications ha2)
			`)
		} else if req.ApplicantEmpty == "0" {
			db = db.Where(`
				hosts.id IN (SELECT ha.host_id FROM host_applications ha WHERE COALESCE(ha.applicant, '') != '')
			`)
		}
	}

	var total int64
	db.Count(&total)

	var hosts []models.Host
	offset := (req.Page - 1) * req.PageSize
	result := db.Preload("Application").Offset(offset).Limit(req.PageSize).Order("id ASC").Find(&hosts)
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
	result := database.DB.Preload("Application").First(&host, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &host, nil
}

type CreateHostRequest struct {
	Region        string `json:"region" binding:"required"`
	InstanceID    string `json:"instance_id"`
	Name          string `json:"name" binding:"required"`
	PrivateIP     string `json:"private_ip" binding:"required"`
	PublicIP      string `json:"public_ip"`
	AssetType     string `json:"asset_type"`
	OS            string `json:"os"`
	CPU           int    `json:"cpu"`
	CPUArch       string `json:"cpu_arch"`
	Memory        int    `json:"memory"`
	Disk          int    `json:"disk"`
	SystemDisk    int    `json:"system_disk"`
	DataDisk      int    `json:"data_disk"`
	EnvType       string `json:"env_type"`
	IsDBServer    *bool  `json:"is_db_server"`
	Status        string `json:"status"`
	OpenPorts     string `json:"open_ports"`
	Tags          string `json:"tags"`

	ApplyUnit        string `json:"apply_unit"`
	Applicant        string `json:"applicant"`
	ApplicantContact string `json:"applicant_contact"`
	Project          string `json:"project"`
	ApplyReason      string `json:"apply_reason"`
	ApplyConfig      string `json:"apply_config"`
	ApplyTime        string `json:"apply_time"`
	ObjectStorageSize string `json:"object_storage_size"`
	Remark           string `json:"remark"`
}

func (s *HostService) Create(req CreateHostRequest) (uint, error) {
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
		PublicIP:   req.PublicIP,
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
	Region        string `json:"region"`
	InstanceID    string `json:"instance_id"`
	Name          string `json:"name"`
	PrivateIP     string `json:"private_ip"`
	PublicIP      string `json:"public_ip"`
	AssetType     string `json:"asset_type"`
	OS            string `json:"os"`
	CPU           int    `json:"cpu"`
	CPUArch       string `json:"cpu_arch"`
	Memory        int    `json:"memory"`
	Disk          int    `json:"disk"`
	SystemDisk    int    `json:"system_disk"`
	DataDisk      int    `json:"data_disk"`
	EnvType       string `json:"env_type"`
	IsDBServer    *bool  `json:"is_db_server"`
	Status        string `json:"status"`
	OpenPorts     string `json:"open_ports"`
	Tags          string `json:"tags"`

	ApplyUnit        *string `json:"apply_unit"`
	Applicant        *string `json:"applicant"`
	ApplicantContact *string `json:"applicant_contact"`
	Project          *string `json:"project"`
	ApplyReason      *string `json:"apply_reason"`
	ApplyConfig      *string `json:"apply_config"`
	ApplyTime        *string `json:"apply_time"`
	ObjectStorageSize *string `json:"object_storage_size"`
	Remark           *string `json:"remark"`
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
	if req.PublicIP != "" {
		updates["public_ip"] = req.PublicIP
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

	tx := database.DB.Begin()
	tx.Where("host_id = ?", id).Delete(&models.HostApplication{})
	tx.Delete(&host)
	tx.Commit()
	return nil
}

type BatchCreateItem struct {
	Region        string `json:"region" binding:"required"`
	InstanceID    string `json:"instance_id"`
	Name          string `json:"name" binding:"required"`
	PrivateIP     string `json:"private_ip" binding:"required"`
	PublicIP      string `json:"public_ip"`
	AssetType     string `json:"asset_type"`
	OS            string `json:"os"`
	CPU           int    `json:"cpu"`
	CPUArch       string `json:"cpu_arch"`
	Memory        int    `json:"memory"`
	Disk          int    `json:"disk"`
	SystemDisk    int    `json:"system_disk"`
	DataDisk      int    `json:"data_disk"`
	EnvType       string `json:"env_type"`
	IsDBServer    *bool  `json:"is_db_server"`
	Status        string `json:"status"`
	OpenPorts     string `json:"open_ports"`
	Tags          string `json:"tags"`
	Applicant     string `json:"applicant"`
	Project       string `json:"project"`
	ApplyUnit     string `json:"apply_unit"`
	ApplyReason   string `json:"apply_reason"`
	ApplyConfig   string `json:"apply_config"`
	ApplyTime     string `json:"apply_time"`
	ObjectStorageSize string `json:"object_storage_size"`
	Remark        string `json:"remark"`
	ApplicantContact string `json:"applicant_contact"`
}

type BatchCreateRequest struct {
	Hosts []BatchCreateItem `json:"hosts" binding:"required"`
}

type BatchCreateResponse struct {
	Success int `json:"success"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

func (s *HostService) BatchCreate(req BatchCreateRequest) (*BatchCreateResponse, error) {
	resp := &BatchCreateResponse{}

	for _, item := range req.Hosts {
		var count int64
		database.DB.Model(&models.Host{}).Where("private_ip = ?", item.PrivateIP).Count(&count)
		if count > 0 {
			resp.Skipped++
			continue
		}

		req2 := CreateHostRequest{
			Region:             item.Region,
			InstanceID:         item.InstanceID,
			Name:               item.Name,
			PrivateIP:          item.PrivateIP,
			PublicIP:           item.PublicIP,
			AssetType:          item.AssetType,
			OS:                 item.OS,
			CPU:                item.CPU,
			CPUArch:            item.CPUArch,
			Memory:             item.Memory,
			Disk:               item.Disk,
			SystemDisk:         item.SystemDisk,
			DataDisk:           item.DataDisk,
			EnvType:            item.EnvType,
			IsDBServer:         item.IsDBServer,
			Status:             item.Status,
			OpenPorts:          item.OpenPorts,
			Tags:               item.Tags,
			Applicant:          item.Applicant,
			Project:            item.Project,
			ApplyUnit:          item.ApplyUnit,
			ApplyReason:        item.ApplyReason,
			ApplyConfig:        item.ApplyConfig,
			ApplyTime:          item.ApplyTime,
			ObjectStorageSize:  item.ObjectStorageSize,
			Remark:             item.Remark,
			ApplicantContact:   item.ApplicantContact,
		}

		_, err := s.Create(req2)
		if err != nil {
			resp.Errors++
		} else {
			resp.Success++
		}
	}

	return resp, nil
}

type BatchUpdateRequest struct {
	IDS  []uint                 `json:"ids" binding:"required"`
	Data map[string]interface{} `json:"data" binding:"required"`
}

func (s *HostService) BatchUpdate(req BatchUpdateRequest) error {
	if len(req.IDS) == 0 {
		return errors.New("未选择任何主机")
	}

	// Fields that belong to host_applications table
	appFields := map[string]bool{
		"apply_unit": true, "applicant": true, "applicant_contact": true,
		"project": true, "apply_reason": true, "apply_config": true,
		"apply_time": true, "object_storage_size": true, "remark": true,
	}

	// Separate host fields and application fields
	hostUpdates := map[string]interface{}{}
	appUpdates := map[string]interface{}{}
	for k, v := range req.Data {
		if appFields[k] {
			appUpdates[k] = v
		} else {
			hostUpdates[k] = v
		}
	}

	tx := database.DB.Begin()

	if len(hostUpdates) > 0 {
		if err := tx.Model(&models.Host{}).Where("id IN ?", req.IDS).Updates(hostUpdates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(appUpdates) > 0 {
		tx.Model(&models.HostApplication{}).Where("host_id IN ?", req.IDS).Updates(appUpdates)
	}

	tx.Commit()
	return nil
}

func (s *HostService) ParseCSVRowToCreateHost(row []string) (CreateHostRequest, error) {
	if len(row) < 26 {
		return CreateHostRequest{}, fmt.Errorf("列数不足，需要26列，实际%d列", len(row))
	}

	cpu, _ := strconv.Atoi(strings.TrimSpace(row[7]))
	memory, _ := strconv.Atoi(strings.TrimSpace(row[9]))
	systemDisk, _ := strconv.Atoi(strings.TrimSpace(row[10]))
	dataDisk, _ := strconv.Atoi(strings.TrimSpace(row[11]))

	isDB := false
	if strings.TrimSpace(row[13]) == "是" || strings.TrimSpace(row[13]) == "true" || strings.TrimSpace(row[13]) == "1" {
		isDB = true
	}

	return CreateHostRequest{
		Region:             strings.TrimSpace(row[0]),
		InstanceID:         strings.TrimSpace(row[1]),
		Name:               strings.TrimSpace(row[2]),
		PrivateIP:          strings.TrimSpace(row[3]),
		PublicIP:           strings.TrimSpace(row[4]),
		AssetType:          strings.TrimSpace(row[5]),
		OS:                 strings.TrimSpace(row[6]),
		CPU:                cpu,
		CPUArch:            strings.TrimSpace(row[8]),
		Memory:             memory,
		SystemDisk:         systemDisk,
		DataDisk:           dataDisk,
		EnvType:            strings.TrimSpace(row[12]),
		IsDBServer:         &isDB,
		Status:             strings.TrimSpace(row[14]),
		OpenPorts:          strings.TrimSpace(row[15]),
		Tags:               strings.TrimSpace(row[16]),
		ApplyUnit:          strings.TrimSpace(row[17]),
		Applicant:          strings.TrimSpace(row[18]),
		ApplicantContact:   strings.TrimSpace(row[19]),
		Project:            strings.TrimSpace(row[20]),
		ApplyReason:        strings.TrimSpace(row[21]),
		ApplyConfig:        strings.TrimSpace(row[22]),
		ApplyTime:          strings.TrimSpace(row[23]),
		ObjectStorageSize:  strings.TrimSpace(row[24]),
		Remark:             strings.TrimSpace(row[25]),
	}, nil
}

func (s *HostService) ExportToCSVRows(hosts []models.Host) [][]string {
	var rows [][]string
	for _, h := range hosts {
		isDB := "否"
		if h.IsDBServer {
			isDB = "是"
		}
		row := []string{
			h.Region, h.InstanceID, h.Name, h.PrivateIP, h.PublicIP,
			h.AssetType, h.OS, strconv.Itoa(h.CPU), h.CPUArch, strconv.Itoa(h.Memory),
			strconv.Itoa(h.SystemDisk), strconv.Itoa(h.DataDisk), h.EnvType, isDB,
			h.Status, h.OpenPorts, h.Tags,
		}
		if h.Application != nil {
			row = append(row,
				h.Application.ApplyUnit, h.Application.Applicant, h.Application.ApplicantContact,
				h.Application.Project, h.Application.ApplyReason, h.Application.ApplyConfig,
				h.Application.ApplyTime, h.Application.ObjectStorageSize, h.Application.Remark,
			)
		} else {
			row = append(row, "", "", "", "", "", "", "", "", "")
		}
		rows = append(rows, row)
	}
	return rows
}
