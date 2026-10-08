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

type PortMappingService struct{}

func NewPortMappingService() *PortMappingService {
	return &PortMappingService{}
}

type PortMappingRequest struct {
	PublicIP      string `json:"public_ip" binding:"required"`
	HostID        uint   `json:"host_id" binding:"required"`
	ExternalPorts string `json:"external_ports" binding:"required"`
	InternalPorts string `json:"internal_ports" binding:"required"`
	Domain        string `json:"domain"`
	ISP           string `json:"isp"`
	ExitLocation  string `json:"exit_location"`
	Remark        string `json:"remark"`
}

var ErrPortMappingNotFound = errors.New("映射记录不存在")

// ErrHostReferencedByMapping 主机已被映射台账引用
var ErrHostReferencedByMapping = errors.New("该主机已被映射台账引用")

// ErrMappingHostNotFound 内网主机不存在
var ErrMappingHostNotFound = errors.New("内网主机不存在")

func parsePortList(field, raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("%s不能为空", field)
	}
	parts := strings.Split(raw, ",")
	ports := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("%s包含不合法端口 %q（范围 1-65535）", field, p)
		}
		ports = append(ports, strconv.Itoa(n))
	}
	return ports, nil
}

func normalizePortMapping(req PortMappingRequest) (PortMappingRequest, []string, []string, error) {
	req.PublicIP = strings.TrimSpace(req.PublicIP)
	req.Domain = strings.TrimSpace(req.Domain)
	req.ISP = strings.TrimSpace(req.ISP)
	req.ExitLocation = strings.TrimSpace(req.ExitLocation)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.PublicIP == "" {
		return req, nil, nil, errors.New("公网IP不能为空")
	}
	if req.HostID == 0 {
		return req, nil, nil, errors.New("内网主机不能为空")
	}
	ext, err := parsePortList("外网端口", req.ExternalPorts)
	if err != nil {
		return req, nil, nil, err
	}
	intl, err := parsePortList("内网端口", req.InternalPorts)
	if err != nil {
		return req, nil, nil, err
	}
	if len(ext) != len(intl) {
		return req, nil, nil, fmt.Errorf("外网端口与内网端口数量必须一致（外网 %d 个，内网 %d 个）", len(ext), len(intl))
	}
	req.ExternalPorts = strings.Join(ext, ",")
	req.InternalPorts = strings.Join(intl, ",")
	return req, ext, intl, nil
}

func refreshHostIPMapped(hostID uint) error {
	if hostID == 0 {
		return nil
	}
	var count int64
	if err := database.DB.Model(&models.PortMapping{}).Where("host_id = ?", hostID).Count(&count).Error; err != nil {
		return err
	}
	return database.DB.Model(&models.Host{}).Where("id = ?", hostID).Update("ip_mapped", count > 0).Error
}

// HostReferencedByMapping 主机是否被映射台账引用（供 host_service 删除前校验）
func HostReferencedByMapping(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.PortMapping{}).Where("host_id = ?", hostID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询映射台账引用失败: %w", err)
	}
	return count, nil
}

func (s *PortMappingService) List(keyword string) ([]models.PortMapping, error) {
	db := database.DB.Model(&models.PortMapping{}).Preload("Host")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(
			"port_mappings.public_ip ILIKE ? OR port_mappings.domain ILIKE ? OR port_mappings.isp ILIKE ? OR port_mappings.exit_location ILIKE ? OR port_mappings.remark ILIKE ? OR port_mappings.external_ports ILIKE ? OR port_mappings.internal_ports ILIKE ? OR EXISTS (SELECT 1 FROM hosts WHERE hosts.id = port_mappings.host_id AND (hosts.name ILIKE ? OR hosts.private_ip ILIKE ?))",
			like, like, like, like, like, like, like, like, like,
		)
	}
	var items []models.PortMapping
	if err := db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PortMappingService) Create(req PortMappingRequest) (uint, error) {
	req, _, _, err := normalizePortMapping(req)
	if err != nil {
		return 0, err
	}
	var hostCount int64
	database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
	if hostCount == 0 {
		return 0, ErrMappingHostNotFound
	}
	if req.Domain != "" {
		var count int64
		database.DB.Model(&models.PortMapping{}).Where("domain = ?", req.Domain).Count(&count)
		if count > 0 {
			return 0, errors.New("该域名已存在")
		}
	}
	rec := models.PortMapping{
		PublicIP:      req.PublicIP,
		HostID:        req.HostID,
		ExternalPorts: req.ExternalPorts,
		InternalPorts: req.InternalPorts,
		Domain:        req.Domain,
		ISP:           req.ISP,
		ExitLocation:  req.ExitLocation,
		Remark:        req.Remark,
	}
	if err := database.DB.Create(&rec).Error; err != nil {
		return 0, err
	}
	if err := refreshHostIPMapped(req.HostID); err != nil {
		return 0, err
	}
	return rec.ID, nil
}

func (s *PortMappingService) Update(id uint, req PortMappingRequest) error {
	req, _, _, err := normalizePortMapping(req)
	if err != nil {
		return err
	}
	var rec models.PortMapping
	if err := database.DB.First(&rec, id).Error; err != nil {
		return ErrPortMappingNotFound
	}
	oldHostID := rec.HostID
	var hostCount int64
	database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
	if hostCount == 0 {
		return ErrMappingHostNotFound
	}
	if req.Domain != "" {
		var count int64
		database.DB.Model(&models.PortMapping{}).Where("domain = ? AND id != ?", req.Domain, id).Count(&count)
		if count > 0 {
			return errors.New("该域名已存在")
		}
	}
	updates := map[string]interface{}{
		"public_ip":      req.PublicIP,
		"host_id":        req.HostID,
		"external_ports": req.ExternalPorts,
		"internal_ports": req.InternalPorts,
		"isp":            req.ISP,
		"exit_location":  req.ExitLocation,
		"remark":         req.Remark,
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	} else {
		updates["domain"] = nil
	}
	if err := database.DB.Model(&rec).Updates(updates).Error; err != nil {
		return err
	}
	if err := refreshHostIPMapped(oldHostID); err != nil {
		return err
	}
	if oldHostID != req.HostID {
		return refreshHostIPMapped(req.HostID)
	}
	return nil
}

func (s *PortMappingService) Delete(id uint) error {
	var rec models.PortMapping
	if err := database.DB.First(&rec, id).Error; err != nil {
		return ErrPortMappingNotFound
	}
	hostID := rec.HostID
	if err := database.DB.Delete(&rec).Error; err != nil {
		return err
	}
	return refreshHostIPMapped(hostID)
}

// BatchCreateText 映射批量添加：公网IP,内网IP,外网端口,内网端口,域名,运营商,出口位置,备注
func (s *PortMappingService) BatchCreateText(text string) (*BatchCreateResponse, error) {
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
		if !headerSkipped && len(row) > 0 && strings.TrimSpace(row[0]) == "公网IP" {
			headerSkipped = true
			continue
		}
		if len(row) > 8 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 列数超出 8 列", lineNo))
			continue
		}
		if len(row) < 4 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 至少需要4列(公网IP,内网IP,外网端口,内网端口)", lineNo))
			continue
		}
		for len(row) < 8 {
			row = append(row, "")
		}
		privateIP := strings.TrimSpace(row[1])
		var host models.Host
		if err := database.DB.Where("private_ip = ?", privateIP).First(&host).Error; err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网主机不存在(%s)", lineNo, privateIP))
			continue
		}
		req := PortMappingRequest{
			PublicIP:      strings.TrimSpace(row[0]),
			HostID:        host.ID,
			ExternalPorts: strings.TrimSpace(row[2]),
			InternalPorts: strings.TrimSpace(row[3]),
			Domain:        strings.TrimSpace(row[4]),
			ISP:           strings.TrimSpace(row[5]),
			ExitLocation:  strings.TrimSpace(row[6]),
			Remark:        strings.TrimSpace(row[7]),
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
