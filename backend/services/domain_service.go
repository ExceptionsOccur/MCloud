package services

import (
	"errors"
	"strings"
	"time"

	"mcloud/database"
	"mcloud/models"
)

type DomainService struct{}

func NewDomainService() *DomainService {
	return &DomainService{}
}

type DomainRequest struct {
	Domain    string     `json:"domain" binding:"required"`
	PublicIP  string     `json:"public_ip"`
	Provider  string     `json:"provider"`
	ExpiresAt *time.Time `json:"expires_at"`
	Remark    string     `json:"remark"`
}

var ErrDomainNotFound = errors.New("域名记录不存在")

func normalizeDomain(req DomainRequest) (DomainRequest, error) {
	req.Domain = strings.TrimSpace(req.Domain)
	req.PublicIP = strings.TrimSpace(req.PublicIP)
	req.Provider = strings.TrimSpace(req.Provider)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Domain == "" {
		return req, errors.New("域名不能为空")
	}
	return req, nil
}

func (s *DomainService) List(keyword string) ([]models.Domain, error) {
	db := database.DB.Model(&models.Domain{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("domain ILIKE ? OR public_ip ILIKE ? OR provider ILIKE ? OR remark ILIKE ?", like, like, like, like)
	}
	var domains []models.Domain
	if err := db.Order("id ASC").Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *DomainService) Create(req DomainRequest) (uint, error) {
	req, err := normalizeDomain(req)
	if err != nil {
		return 0, err
	}
	var count int64
	database.DB.Model(&models.Domain{}).Where("domain = ?", req.Domain).Count(&count)
	if count > 0 {
		return 0, errors.New("该域名已存在")
	}
	d := models.Domain{
		Domain:    req.Domain,
		PublicIP:  req.PublicIP,
		Provider:  req.Provider,
		ExpiresAt: req.ExpiresAt,
		Remark:    req.Remark,
	}
	if err := database.DB.Create(&d).Error; err != nil {
		return 0, err
	}
	return d.ID, nil
}

func (s *DomainService) Update(id uint, req DomainRequest) error {
	req, err := normalizeDomain(req)
	if err != nil {
		return err
	}
	var d models.Domain
	if err := database.DB.First(&d, id).Error; err != nil {
		return ErrDomainNotFound
	}
	var count int64
	database.DB.Model(&models.Domain{}).Where("domain = ? AND id != ?", req.Domain, id).Count(&count)
	if count > 0 {
		return errors.New("该域名已存在")
	}
	d.Domain = req.Domain
	d.PublicIP = req.PublicIP
	d.Provider = req.Provider
	d.ExpiresAt = req.ExpiresAt
	d.Remark = req.Remark
	return database.DB.Save(&d).Error
}

func (s *DomainService) Delete(id uint) error {
	var d models.Domain
	if err := database.DB.First(&d, id).Error; err != nil {
		return ErrDomainNotFound
	}
	return database.DB.Delete(&d).Error
}
