package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

type DomainService struct{}

func NewDomainService() *DomainService {
	return &DomainService{}
}

type DomainRequest struct {
	Domain   string `json:"domain" binding:"required"`
	PublicIP string `json:"public_ip"`
	ISP      string `json:"isp"`
	HostID   uint   `json:"host_id"`
	HostPort int    `json:"host_port"`
	Remark   string `json:"remark"`
}

var ErrDomainNotFound = errors.New("域名记录不存在")

// ErrHostNotFound 主机不存在
var ErrDomainHostNotFound = errors.New("内网主机不存在")

// ErrHostReferencedByDomain 主机已被域名台账引用
var ErrHostReferencedByDomain = errors.New("该主机已被域名台账引用")

func normalizeDomain(req DomainRequest) (DomainRequest, error) {
	req.Domain = strings.TrimSpace(req.Domain)
	req.PublicIP = strings.TrimSpace(req.PublicIP)
	req.ISP = strings.TrimSpace(req.ISP)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Domain == "" {
		return req, errors.New("域名不能为空")
	}
	if req.HostPort < 0 || req.HostPort > 65535 {
		return req, errors.New("主机端口必须在 0-65535 之间")
	}
	return req, nil
}

func (s *DomainService) List(keyword string) ([]models.Domain, error) {
	db := database.DB.Model(&models.Domain{}).Preload("Host")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(
			"domains.domain ILIKE ? OR domains.public_ip ILIKE ? OR domains.isp ILIKE ? OR domains.remark ILIKE ? OR EXISTS (SELECT 1 FROM hosts WHERE hosts.id = domains.host_id AND (hosts.name ILIKE ? OR hosts.private_ip ILIKE ?))",
			like, like, like, like, like, like,
		)
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
	if req.HostID > 0 {
		var hostCount int64
		database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
		if hostCount == 0 {
			return 0, ErrDomainHostNotFound
		}
	}
	var count int64
	database.DB.Model(&models.Domain{}).Where("domain = ?", req.Domain).Count(&count)
	if count > 0 {
		return 0, errors.New("该域名已存在")
	}
	d := models.Domain{
		Domain:   req.Domain,
		PublicIP: req.PublicIP,
		ISP:      req.ISP,
		HostID:   req.HostID,
		HostPort: req.HostPort,
		Remark:   req.Remark,
	}
	cols := []string{"domain", "public_ip", "isp", "host_port", "remark"}
	if req.HostID > 0 {
		cols = append(cols, "host_id")
	}
	if err := database.DB.Select(cols).Create(&d).Error; err != nil {
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
	if req.HostID > 0 {
		var hostCount int64
		database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
		if hostCount == 0 {
			return ErrDomainHostNotFound
		}
	}
	var count int64
	database.DB.Model(&models.Domain{}).Where("domain = ? AND id != ?", req.Domain, id).Count(&count)
	if count > 0 {
		return errors.New("该域名已存在")
	}
	d.Domain = req.Domain
	d.PublicIP = req.PublicIP
	d.ISP = req.ISP
	d.HostPort = req.HostPort
	d.Remark = req.Remark
	if req.HostID > 0 {
		d.HostID = req.HostID
		return database.DB.Save(&d).Error
	}
	// 清除关联：显式置空 host_id，避免写入 0 违反外键
	return database.DB.Model(&d).Updates(map[string]interface{}{
		"domain":    d.Domain,
		"public_ip": d.PublicIP,
		"isp":       d.ISP,
		"host_id":   nil,
		"host_port": d.HostPort,
		"remark":    d.Remark,
	}).Error
}

func (s *DomainService) Delete(id uint) error {
	var d models.Domain
	if err := database.DB.First(&d, id).Error; err != nil {
		return ErrDomainNotFound
	}
	return database.DB.Delete(&d).Error
}

// HostReferencedByDomain 主机是否被域名台账引用（供 host_service 删除前校验）
func HostReferencedByDomain(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Domain{}).Where("host_id = ?", hostID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询域名台账引用失败: %w", err)
	}
	return count, nil
}

// BatchCreateText 域名批量添加：列顺序 域名,解析公网IP,运营商,内网IP,主机端口,备注
// 域名已存在则跳过；内网IP为空表示不关联主机；填了内网IP但主机不存在则该行失败
func (s *DomainService) BatchCreateText(text string) (*BatchCreateResponse, error) {
	resp := &BatchCreateResponse{}
	headerSkipped := false

	for i, raw := range strings.Split(text, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		row, err := utils.ParseCSVLine(line)
		if err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 格式错误(%v)", lineNo, err))
			continue
		}

		if !headerSkipped && len(row) > 0 && strings.TrimSpace(row[0]) == "域名" {
			headerSkipped = true
			continue
		}

		if len(row) > 6 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 列数超出 6 列", lineNo))
			continue
		}
		if len(row) < 1 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 至少需要1列(域名)", lineNo))
			continue
		}
		for len(row) < 6 {
			row = append(row, "")
		}

		domainName := strings.TrimSpace(row[0])
		if domainName == "" {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 域名不能为空", lineNo))
			continue
		}

		var count int64
		database.DB.Model(&models.Domain{}).Where("domain = ?", domainName).Count(&count)
		if count > 0 {
			resp.Skipped++
			continue
		}

		req := DomainRequest{
			Domain:   domainName,
			PublicIP: strings.TrimSpace(row[1]),
			ISP:      strings.TrimSpace(row[2]),
			Remark:   strings.TrimSpace(row[5]),
		}
		if privateIP := strings.TrimSpace(row[3]); privateIP != "" {
			var host models.Host
			if err := database.DB.Where("private_ip = ?", privateIP).First(&host).Error; err != nil {
				resp.Errors++
				resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网主机不存在(%s)", lineNo, privateIP))
				continue
			}
			req.HostID = host.ID
			if p := strings.TrimSpace(row[4]); p != "" {
				port, err := strconv.Atoi(p)
				if err != nil || port < 0 || port > 65535 {
					resp.Errors++
					resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 主机端口必须在 0-65535 之间", lineNo))
					continue
				}
				req.HostPort = port
			}
		}

		if _, err := s.Create(req); err != nil {
			if strings.Contains(err.Error(), "已存在") {
				resp.Skipped++
			} else {
				resp.Errors++
				resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
			}
		} else {
			resp.Success++
		}
	}

	return resp, nil
}
