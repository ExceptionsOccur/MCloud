package services

import (
	"errors"
	"net"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

type PublicIPService struct{}

func NewPublicIPService() *PublicIPService {
	return &PublicIPService{}
}

type PublicIPRequest struct {
	IP           string `json:"ip" binding:"required"`
	ISP          string `json:"isp"`
	ExitLocation string `json:"exit_location"`
	Remark       string `json:"remark"`
}

var ErrPublicIPNotFound = errors.New("公网IP记录不存在")

// ErrPublicIPReferenced 公网IP已被映射台账或零信任台账引用，禁止删除/修改
var ErrPublicIPReferenced = errors.New("该公网IP已被映射台账或零信任台账引用")

// countPublicIPRefs 统计公网IP在 port_mappings / zero_trusts 中的引用数
func countPublicIPRefs(ip string) (int64, error) {
	var pm int64
	if err := database.DB.Model(&models.PortMapping{}).Where("public_ip = ?", ip).Count(&pm).Error; err != nil {
		return 0, err
	}
	var zt int64
	if err := database.DB.Model(&models.ZeroTrust{}).Where("public_ip = ?", ip).Count(&zt).Error; err != nil {
		return 0, err
	}
	return pm + zt, nil
}

func normalizePublicIP(req PublicIPRequest) (PublicIPRequest, error) {
	req.IP = strings.TrimSpace(req.IP)
	req.ISP = strings.TrimSpace(req.ISP)
	req.ExitLocation = strings.TrimSpace(req.ExitLocation)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.IP == "" {
		return req, errors.New("公网IP不能为空")
	}
	if net.ParseIP(req.IP) == nil {
		return req, errors.New("公网IP格式不正确")
	}
	return req, nil
}

func (s *PublicIPService) List(keyword string) ([]models.PublicIP, error) {
	db := database.DB.Model(&models.PublicIP{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("ip ILIKE ? OR isp ILIKE ? OR exit_location ILIKE ? OR remark ILIKE ?", like, like, like, like)
	}
	var items []models.PublicIP
	if err := db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PublicIPService) Create(req PublicIPRequest, op Operator, requestID string) (uint, error) {
	req, err := normalizePublicIP(req)
	if err != nil {
		return 0, err
	}
	var count int64
	database.DB.Model(&models.PublicIP{}).Where("ip = ?", req.IP).Count(&count)
	if count > 0 {
		return 0, errors.New("该公网IP已存在")
	}
	record := models.PublicIP{IP: req.IP, ISP: req.ISP, ExitLocation: req.ExitLocation, Remark: req.Remark}
	if err := database.DB.Create(&record).Error; err != nil {
		return 0, err
	}
	NewAuditService().Record(op, "create", "public_ip", &record.ID, DetailDiff{After: req}, requestID)
	return record.ID, nil
}

func (s *PublicIPService) Update(id uint, req PublicIPRequest, op Operator, requestID string) error {
	req, err := normalizePublicIP(req)
	if err != nil {
		return err
	}
	var record models.PublicIP
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrPublicIPNotFound
	}
	var count int64
	database.DB.Model(&models.PublicIP{}).Where("ip = ? AND id != ?", req.IP, id).Count(&count)
	if count > 0 {
		return errors.New("该公网IP已存在")
	}
	before := PublicIPRequest{IP: record.IP, ISP: record.ISP, ExitLocation: record.ExitLocation, Remark: record.Remark}
	if req.IP != record.IP {
		n, err := countPublicIPRefs(record.IP)
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrPublicIPReferenced
		}
	}
	record.IP = req.IP
	record.ISP = req.ISP
	record.ExitLocation = req.ExitLocation
	record.Remark = req.Remark
	if err := database.DB.Save(&record).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "update", "public_ip", &id, DetailDiff{Before: before, After: req}, requestID)
	return nil
}

func (s *PublicIPService) Delete(id uint, op Operator, requestID string) error {
	var record models.PublicIP
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrPublicIPNotFound
	}
	n, err := countPublicIPRefs(record.IP)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrPublicIPReferenced
	}
	if err := database.DB.Delete(&record).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "delete", "public_ip", &id,
		DetailDiff{Before: PublicIPRequest{IP: record.IP, ISP: record.ISP, ExitLocation: record.ExitLocation, Remark: record.Remark}}, requestID)
	return nil
}
