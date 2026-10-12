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
	PublicIP          string `json:"public_ip" binding:"required"`
	HostID            uint   `json:"host_id" binding:"required"`
	ExternalPorts     string `json:"external_ports" binding:"required"`
	ExternalProtocols string `json:"external_protocols"`
	InternalPorts     string `json:"internal_ports" binding:"required"`
	InternalProtocols string `json:"internal_protocols"`
	Domain            string `json:"domain"`
	Remark            string `json:"remark"`
}

var ErrPortMappingNotFound = errors.New("映射记录不存在")

// ErrPortMappingExists 完全相同的映射（host_id + public_ip + external_ports + external_protocols 整组）已存在
var ErrPortMappingExists = errors.New("该映射已存在")

// ErrHostReferencedByMapping 主机已被映射台账引用
var ErrHostReferencedByMapping = errors.New("该主机已被映射台账引用")

// ErrMappingHostNotFound 内网主机不存在
var ErrMappingHostNotFound = errors.New("内网主机不存在")

// parsePortExpr 解析端口表达式（单端口或 A-B 连续范围，T-055）：
// 返回规范化 token（数字与范围两端均去前导零），不展开、不去重
func parsePortExpr(field, raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("%s不能为空", field)
	}
	parts := strings.Split(raw, ",")
	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("%s包含空端口项", field)
		}
		if strings.Contains(p, "-") {
			seg := strings.Split(p, "-")
			if len(seg) != 2 {
				return nil, fmt.Errorf("%s包含不合法端口 %q（范围 1-65535，连续区间写作起始-结束）", field, p)
			}
			a, errA := strconv.Atoi(strings.TrimSpace(seg[0]))
			b, errB := strconv.Atoi(strings.TrimSpace(seg[1]))
			if errA != nil || errB != nil || a < 1 || b > 65535 || a > b {
				return nil, fmt.Errorf("%s包含不合法端口范围 %q（1-65535 且起始 ≤ 结束）", field, p)
			}
			tokens = append(tokens, strconv.Itoa(a)+"-"+strconv.Itoa(b))
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("%s包含不合法端口 %q（范围 1-65535）", field, p)
		}
		tokens = append(tokens, strconv.Itoa(n))
	}
	return tokens, nil
}

// expandPortCount 端口 token 展开后的总端口数（单口 1，A-B 为 B-A+1）
func expandPortCount(tokens []string) int {
	total := 0
	for _, t := range tokens {
		if i := strings.IndexByte(t, '-'); i >= 0 {
			a, errA := strconv.Atoi(t[:i])
			b, errB := strconv.Atoi(t[i+1:])
			if errA == nil && errB == nil && b >= a {
				total += b - a + 1
				continue
			}
		}
		total++
	}
	return total
}

// parseProtocols 校验与端口 token 等长并行的传输协议列表（仅 tcp/udp，统一小写）；
// raw 为空时按 tcp 缺省补齐 n 个（n 为 token 数，范围 token 计 1）（T-055）
func parseProtocols(field, raw string, n int) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		protos := make([]string, n)
		for i := range protos {
			protos[i] = "tcp"
		}
		return protos, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != n {
		return nil, fmt.Errorf("%s数量必须与端口数量一致（协议 %d 个，端口 %d 个）", field, len(parts), n)
	}
	protos := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "tcp" && p != "udp" {
			return nil, fmt.Errorf("%s包含不合法协议 %q（仅 tcp/udp）", field, p)
		}
		protos = append(protos, p)
	}
	return protos, nil
}

func normalizePortMapping(req PortMappingRequest) (PortMappingRequest, []string, []string, error) {
	req.PublicIP = strings.TrimSpace(req.PublicIP)
	req.Domain = strings.TrimSpace(req.Domain)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.PublicIP == "" {
		return req, nil, nil, errors.New("公网IP不能为空")
	}
	if req.HostID == 0 {
		return req, nil, nil, errors.New("内网主机不能为空")
	}
	ext, err := parsePortExpr("外网端口", req.ExternalPorts)
	if err != nil {
		return req, nil, nil, err
	}
	intl, err := parsePortExpr("内网端口", req.InternalPorts)
	if err != nil {
		return req, nil, nil, err
	}
	if extCount, intlCount := expandPortCount(ext), expandPortCount(intl); extCount != intlCount {
		return req, nil, nil, fmt.Errorf("外网端口与内网端口数量必须一致（外网 %d 个，内网 %d 个）", extCount, intlCount)
	}
	extProtos, err := parseProtocols("外网协议", req.ExternalProtocols, len(ext))
	if err != nil {
		return req, nil, nil, err
	}
	intlProtos, err := parseProtocols("内网协议", req.InternalProtocols, len(intl))
	if err != nil {
		return req, nil, nil, err
	}
	req.ExternalPorts = strings.Join(ext, ",")
	req.InternalPorts = strings.Join(intl, ",")
	req.ExternalProtocols = strings.Join(extProtos, ",")
	req.InternalProtocols = strings.Join(intlProtos, ",")
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
			"port_mappings.public_ip ILIKE ? OR port_mappings.domain ILIKE ? OR port_mappings.remark ILIKE ? OR port_mappings.external_ports ILIKE ? OR port_mappings.internal_ports ILIKE ? OR port_mappings.external_protocols ILIKE ? OR port_mappings.internal_protocols ILIKE ? OR EXISTS (SELECT 1 FROM hosts WHERE hosts.id = port_mappings.host_id AND (hosts.name ILIKE ? OR hosts.private_ip ILIKE ?)) OR EXISTS (SELECT 1 FROM public_ips pi WHERE pi.ip = port_mappings.public_ip AND (pi.isp ILIKE ? OR pi.exit_location ILIKE ?))",
			like, like, like, like, like, like, like, like, like, like, like,
		)
	}
	var items []models.PortMapping
	if err := db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	// 运营商/出口位置为派生字段（本表不存储），恒从公网IP资源池带出（对齐 zero_trusts 做法）
	if len(items) > 0 {
		ipSet := make(map[string]struct{}, len(items))
		ips := make([]string, 0, len(items))
		for _, it := range items {
			if it.PublicIP == "" {
				continue
			}
			if _, ok := ipSet[it.PublicIP]; !ok {
				ipSet[it.PublicIP] = struct{}{}
				ips = append(ips, it.PublicIP)
			}
		}
		if len(ips) > 0 {
			var resources []models.PublicIP
			if err := database.DB.Where("ip IN ?", ips).Find(&resources).Error; err == nil {
				byIP := make(map[string]models.PublicIP, len(resources))
				for _, r := range resources {
					byIP[r.IP] = r
				}
				for i := range items {
					if r, ok := byIP[items[i].PublicIP]; ok {
						items[i].ISP = r.ISP
						items[i].ExitLocation = r.ExitLocation
					}
				}
			}
		}
	}
	return items, nil
}

func (s *PortMappingService) Create(req PortMappingRequest, op Operator, requestID string) (uint, error) {
	req, _, _, err := normalizePortMapping(req)
	if err != nil {
		return 0, err
	}
	var hostCount int64
	database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
	if hostCount == 0 {
		return 0, ErrMappingHostNotFound
	}
	var ipCount int64
	database.DB.Model(&models.PublicIP{}).Where("ip = ?", req.PublicIP).Count(&ipCount)
	if ipCount == 0 {
		return 0, errors.New("公网IP不在资源池中，请先在公网IP录入中添加")
	}
	if req.Domain != "" {
		var count int64
		database.DB.Model(&models.PortMapping{}).Where("domain = ?", req.Domain).Count(&count)
		if count > 0 {
			return 0, errors.New("该域名已存在")
		}
	}
	var dupCount int64
	database.DB.Model(&models.PortMapping{}).
		Where("host_id = ? AND public_ip = ? AND external_ports = ? AND external_protocols = ?", req.HostID, req.PublicIP, req.ExternalPorts, req.ExternalProtocols).
		Count(&dupCount)
	if dupCount > 0 {
		return 0, ErrPortMappingExists
	}
	rec := models.PortMapping{
		PublicIP:          req.PublicIP,
		HostID:            req.HostID,
		ExternalPorts:     req.ExternalPorts,
		ExternalProtocols: req.ExternalProtocols,
		InternalPorts:     req.InternalPorts,
		InternalProtocols: req.InternalProtocols,
		Domain:            req.Domain,
		Remark:            req.Remark,
	}
	if err := database.DB.Create(&rec).Error; err != nil {
		return 0, err
	}
	NewAuditService().Record(op, "create", "port_mapping", &rec.ID, DetailDiff{After: req}, requestID)
	if err := refreshHostIPMapped(req.HostID); err != nil {
		return 0, err
	}
	return rec.ID, nil
}

func (s *PortMappingService) Update(id uint, req PortMappingRequest, op Operator, requestID string) error {
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
	var ipCount int64
	database.DB.Model(&models.PublicIP{}).Where("ip = ?", req.PublicIP).Count(&ipCount)
	if ipCount == 0 {
		return errors.New("公网IP不在资源池中，请先在公网IP录入中添加")
	}
	if req.Domain != "" {
		var count int64
		database.DB.Model(&models.PortMapping{}).Where("domain = ? AND id != ?", req.Domain, id).Count(&count)
		if count > 0 {
			return errors.New("该域名已存在")
		}
	}
	var dupCount int64
	database.DB.Model(&models.PortMapping{}).
		Where("host_id = ? AND public_ip = ? AND external_ports = ? AND external_protocols = ? AND id != ?", req.HostID, req.PublicIP, req.ExternalPorts, req.ExternalProtocols, id).
		Count(&dupCount)
	if dupCount > 0 {
		return ErrPortMappingExists
	}
	before := PortMappingRequest{
		PublicIP: rec.PublicIP, HostID: rec.HostID, ExternalPorts: rec.ExternalPorts,
		ExternalProtocols: rec.ExternalProtocols, InternalPorts: rec.InternalPorts,
		InternalProtocols: rec.InternalProtocols, Domain: rec.Domain, Remark: rec.Remark,
	}
	updates := map[string]interface{}{
		"public_ip":          req.PublicIP,
		"host_id":            req.HostID,
		"external_ports":     req.ExternalPorts,
		"external_protocols": req.ExternalProtocols,
		"internal_ports":     req.InternalPorts,
		"internal_protocols": req.InternalProtocols,
		"remark":             req.Remark,
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	} else {
		updates["domain"] = nil
	}
	if err := database.DB.Model(&rec).Updates(updates).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "update", "port_mapping", &id, DetailDiff{Before: before, After: req}, requestID)
	if err := refreshHostIPMapped(oldHostID); err != nil {
		return err
	}
	if oldHostID != req.HostID {
		return refreshHostIPMapped(req.HostID)
	}
	return nil
}

func (s *PortMappingService) Delete(id uint, op Operator, requestID string) error {
	var rec models.PortMapping
	if err := database.DB.First(&rec, id).Error; err != nil {
		return ErrPortMappingNotFound
	}
	hostID := rec.HostID
	if err := database.DB.Delete(&rec).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "delete", "port_mapping", &id,
		DetailDiff{Before: PortMappingRequest{
			PublicIP: rec.PublicIP, HostID: rec.HostID, ExternalPorts: rec.ExternalPorts,
			ExternalProtocols: rec.ExternalProtocols, InternalPorts: rec.InternalPorts,
			InternalProtocols: rec.InternalProtocols, Domain: rec.Domain, Remark: rec.Remark,
		}}, requestID)
	return refreshHostIPMapped(hostID)
}

// BatchCreateText 映射批量添加：公网IP,内网IP,外网端口,内网端口,域名,备注,外网协议,内网协议
// 外网协议/内网协议为第 7/8 可选列，缺省 tcp（仅 tcp/udp）；
// 运营商/出口位置从公网IP资源池展示，不在批量行填写
func (s *PortMappingService) BatchCreateText(text string, op Operator, requestID string) (*BatchCreateResponse, error) {
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
			PublicIP:          strings.TrimSpace(row[0]),
			HostID:            host.ID,
			ExternalPorts:     strings.TrimSpace(row[2]),
			InternalPorts:     strings.TrimSpace(row[3]),
			Domain:            strings.TrimSpace(row[4]),
			Remark:            strings.TrimSpace(row[5]),
			ExternalProtocols: strings.TrimSpace(row[6]),
			InternalProtocols: strings.TrimSpace(row[7]),
		}
		if _, err := s.Create(req, op, requestID); err != nil {
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
