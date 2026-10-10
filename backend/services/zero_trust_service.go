package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

type ZeroTrustService struct{}

func NewZeroTrustService() *ZeroTrustService {
	return &ZeroTrustService{}
}

// ZeroTrustTarget 一组「主机:端口」配对
type ZeroTrustTarget struct {
	HostID uint `json:"host_id"`
	Port   int  `json:"port"`
}

type ZeroTrustRequest struct {
	ApplyUnit   string              `json:"apply_unit" binding:"required"`
	AccountName string              `json:"account_name" binding:"required"`
	Contact     string              `json:"contact"`
	PublicIP    string              `json:"public_ip"`
	Targets     []ZeroTrustTarget   `json:"targets" binding:"required"`
	SystemName  string              `json:"system_name"`
	ApplyTime   *utils.FlexibleTime `json:"apply_time"`
	Remark      string              `json:"remark"`
}

// ZeroTrustTargetView 配对 + 主机简要信息（列表展示用）
type ZeroTrustTargetView struct {
	HostID    uint   `json:"host_id"`
	Port      int    `json:"port"`
	HostName  string `json:"host_name"`
	PrivateIP string `json:"private_ip"`
}

type ZeroTrustItem struct {
	models.ZeroTrust
	Targets      []ZeroTrustTargetView `json:"targets"`
	ExitLocation string                `json:"exit_location"`
}

// ErrZeroTrustNotFound 零信任台账记录不存在
var ErrZeroTrustNotFound = errors.New("零信任台账记录不存在")

// ErrHostNotFound 主机不存在
var ErrHostNotFound = errors.New("主机不存在")

// ErrHostReferenced 主机已被零信任台账引用，禁止删除
var ErrHostReferenced = errors.New("该主机已被零信任台账引用")

// ErrPublicIPNotInPool 公网IP不在资源池中
var ErrPublicIPNotInPool = errors.New("公网IP不在资源池中，请先在公网IP录入中添加")

// ZeroTrustHostBrief 主机简要信息
type ZeroTrustHostBrief struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	PrivateIP string `json:"private_ip"`
}

// normalizeTargetPairs 校验配对列表（端口范围、去重），返回归一化配对
func normalizeTargetPairs(list []ZeroTrustTarget) ([]ZeroTrustTarget, error) {
	if len(list) == 0 {
		return nil, errors.New("接入主机与端口不能为空")
	}
	seen := make(map[ZeroTrustTarget]struct{}, len(list))
	pairs := make([]ZeroTrustTarget, 0, len(list))
	for _, t := range list {
		if t.HostID == 0 {
			return nil, errors.New("申请主机不能为空")
		}
		if t.Port < 1 || t.Port > 65535 {
			return nil, fmt.Errorf("申请端口必须在 1-65535 之间（当前 %d）", t.Port)
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		pairs = append(pairs, t)
	}
	return pairs, nil
}

// joinTargetPairs 配对列表 → 存储串 "host_id:port,..."
func joinTargetPairs(pairs []ZeroTrustTarget) string {
	parts := make([]string, len(pairs))
	for i, t := range pairs {
		parts[i] = strconv.FormatUint(uint64(t.HostID), 10) + ":" + strconv.Itoa(t.Port)
	}
	return strings.Join(parts, ",")
}

// parseStoredTargets 宽松解析存储串（跳过畸形段），供列表/引用回填
func parseStoredTargets(raw string) []ZeroTrustTarget {
	pairs := make([]ZeroTrustTarget, 0)
	for _, seg := range strings.Split(raw, ",") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		idStr, portStr, ok := strings.Cut(seg, ":")
		if !ok {
			continue
		}
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		port, err := strconv.Atoi(strings.TrimSpace(portStr))
		if err != nil || port < 1 || port > 65535 {
			continue
		}
		pairs = append(pairs, ZeroTrustTarget{HostID: uint(id), Port: port})
	}
	return pairs
}

// validateZeroTrustHosts 配对引用的主机必须全部存在
func validateZeroTrustHosts(pairs []ZeroTrustTarget) error {
	idSet := make(map[uint]struct{}, len(pairs))
	ids := make([]uint, 0, len(pairs))
	for _, t := range pairs {
		if _, ok := idSet[t.HostID]; !ok {
			idSet[t.HostID] = struct{}{}
			ids = append(ids, t.HostID)
		}
	}
	var count int64
	if err := database.DB.Model(&models.Host{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if int(count) != len(ids) {
		return ErrHostNotFound
	}
	return nil
}

// validateZeroTrustPublicIP 公网IP非空时必须在资源池中
func validateZeroTrustPublicIP(ip string) error {
	if ip == "" {
		return nil
	}
	var count int64
	if err := database.DB.Model(&models.PublicIP{}).Where("ip = ?", ip).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrPublicIPNotInPool
	}
	return nil
}

// loadZeroTrustHostBriefs 批量加载主机简要信息（ID → 简要）
func loadZeroTrustHostBriefs(idSet map[uint]struct{}) (map[uint]ZeroTrustHostBrief, error) {
	briefs := make(map[uint]ZeroTrustHostBrief, len(idSet))
	if len(idSet) == 0 {
		return briefs, nil
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var hosts []models.Host
	if err := database.DB.Select("id, name, private_ip").Where("id IN ?", ids).Find(&hosts).Error; err != nil {
		return nil, err
	}
	for _, h := range hosts {
		briefs[h.ID] = ZeroTrustHostBrief{ID: h.ID, Name: h.Name, PrivateIP: h.PrivateIP}
	}
	return briefs, nil
}

// loadExitLocations 批量带出公网IP的出口位置（IP → exit_location）
func loadExitLocations(ips []string) (map[string]string, error) {
	locs := make(map[string]string, len(ips))
	if len(ips) == 0 {
		return locs, nil
	}
	var resources []models.PublicIP
	if err := database.DB.Select("ip, exit_location").Where("ip IN ?", ips).Find(&resources).Error; err != nil {
		return nil, err
	}
	for _, r := range resources {
		locs[r.IP] = r.ExitLocation
	}
	return locs, nil
}

func normalizeZeroTrust(req ZeroTrustRequest) (ZeroTrustRequest, []ZeroTrustTarget, error) {
	req.ApplyUnit = strings.TrimSpace(req.ApplyUnit)
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Contact = strings.TrimSpace(req.Contact)
	req.PublicIP = strings.TrimSpace(req.PublicIP)
	req.SystemName = strings.TrimSpace(req.SystemName)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.ApplyUnit == "" {
		return req, nil, errors.New("申请单位不能为空")
	}
	if req.AccountName == "" {
		return req, nil, errors.New("账户名不能为空")
	}
	pairs, err := normalizeTargetPairs(req.Targets)
	if err != nil {
		return req, nil, err
	}
	req.Targets = pairs
	if req.ApplyTime == nil {
		req.ApplyTime = &utils.FlexibleTime{Time: time.Now()}
	}
	return req, pairs, nil
}

func (s *ZeroTrustService) List(keyword string) ([]ZeroTrustItem, error) {
	db := database.DB.Model(&models.ZeroTrust{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(
			"zero_trusts.apply_unit ILIKE ? OR zero_trusts.account_name ILIKE ? OR zero_trusts.contact ILIKE ? OR zero_trusts.system_name ILIKE ? OR zero_trusts.remark ILIKE ? OR zero_trusts.public_ip ILIKE ? OR EXISTS (SELECT 1 FROM public_ips pi WHERE pi.ip = zero_trusts.public_ip AND (pi.exit_location ILIKE ? OR pi.isp ILIKE ?)) OR EXISTS (SELECT 1 FROM hosts WHERE (hosts.name ILIKE ? OR hosts.private_ip ILIKE ?) AND (',' || zero_trusts.targets || ',') LIKE '%,' || hosts.id::text || ':%')",
			like, like, like, like, like, like, like, like, like, like,
		)
	}

	var records []models.ZeroTrust
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	hostIDSet := make(map[uint]struct{})
	parsed := make([][]ZeroTrustTarget, len(records))
	for i, r := range records {
		pairs := parseStoredTargets(r.Targets)
		parsed[i] = pairs
		for _, t := range pairs {
			hostIDSet[t.HostID] = struct{}{}
		}
	}
	briefs, err := loadZeroTrustHostBriefs(hostIDSet)
	if err != nil {
		return nil, err
	}

	ipSet := make(map[string]struct{})
	for _, r := range records {
		if r.PublicIP != "" {
			ipSet[r.PublicIP] = struct{}{}
		}
	}
	ips := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		ips = append(ips, ip)
	}
	locs, err := loadExitLocations(ips)
	if err != nil {
		return nil, err
	}

	items := make([]ZeroTrustItem, 0, len(records))
	for i, r := range records {
		item := ZeroTrustItem{ZeroTrust: r, Targets: make([]ZeroTrustTargetView, 0, len(parsed[i]))}
		for _, t := range parsed[i] {
			view := ZeroTrustTargetView{HostID: t.HostID, Port: t.Port}
			if b, ok := briefs[t.HostID]; ok {
				view.HostName = b.Name
				view.PrivateIP = b.PrivateIP
			}
			item.Targets = append(item.Targets, view)
		}
		item.ExitLocation = locs[r.PublicIP]
		items = append(items, item)
	}
	return items, nil
}

func (s *ZeroTrustService) GetByID(id uint) (*models.ZeroTrust, error) {
	var record models.ZeroTrust
	if err := database.DB.First(&record, id).Error; err != nil {
		return nil, ErrZeroTrustNotFound
	}
	return &record, nil
}

func (s *ZeroTrustService) Create(req ZeroTrustRequest, op Operator, requestID string) (uint, error) {
	req, pairs, err := normalizeZeroTrust(req)
	if err != nil {
		return 0, err
	}
	if err := validateZeroTrustHosts(pairs); err != nil {
		return 0, err
	}
	if err := validateZeroTrustPublicIP(req.PublicIP); err != nil {
		return 0, err
	}

	record := models.ZeroTrust{
		ApplyUnit:   req.ApplyUnit,
		AccountName: req.AccountName,
		Contact:     req.Contact,
		PublicIP:    req.PublicIP,
		Targets:     joinTargetPairs(pairs),
		SystemName:  req.SystemName,
		ApplyTime:   req.ApplyTime.Time,
		Remark:      req.Remark,
	}
	if err := database.DB.Create(&record).Error; err != nil {
		return 0, err
	}
	NewAuditService().Record(op, "create", "zero_trust", &record.ID, DetailDiff{After: req}, requestID)
	return record.ID, nil
}

func (s *ZeroTrustService) Update(id uint, req ZeroTrustRequest, op Operator, requestID string) error {
	req, pairs, err := normalizeZeroTrust(req)
	if err != nil {
		return err
	}

	var record models.ZeroTrust
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrZeroTrustNotFound
	}
	if err := validateZeroTrustHosts(pairs); err != nil {
		return err
	}
	if err := validateZeroTrustPublicIP(req.PublicIP); err != nil {
		return err
	}

	before := map[string]interface{}{
		"apply_unit":   record.ApplyUnit,
		"account_name": record.AccountName,
		"contact":      record.Contact,
		"public_ip":    record.PublicIP,
		"targets":      record.Targets,
		"system_name":  record.SystemName,
		"remark":       record.Remark,
	}
	record.ApplyUnit = req.ApplyUnit
	record.AccountName = req.AccountName
	record.Contact = req.Contact
	record.PublicIP = req.PublicIP
	record.Targets = joinTargetPairs(pairs)
	record.SystemName = req.SystemName
	record.ApplyTime = req.ApplyTime.Time
	record.Remark = req.Remark
	if err := database.DB.Save(&record).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "update", "zero_trust", &id, DetailDiff{Before: before, After: req}, requestID)
	return nil
}

func (s *ZeroTrustService) Delete(id uint, op Operator, requestID string) error {
	var record models.ZeroTrust
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrZeroTrustNotFound
	}
	if err := database.DB.Delete(&record).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "delete", "zero_trust", &id,
		DetailDiff{Before: map[string]interface{}{
			"apply_unit": record.ApplyUnit, "account_name": record.AccountName, "contact": record.Contact,
			"public_ip": record.PublicIP, "targets": record.Targets, "system_name": record.SystemName, "remark": record.Remark,
		}}, requestID)
	return nil
}

// hostInTargetsCond 主机ID出现在 targets 配对中的匹配表达式（防子串误匹配，两侧补逗号）
const hostInTargetsCond = "(',' || targets || ',') LIKE ?"

func hostInTargetsPattern(hostID uint) string {
	return "%," + strconv.FormatUint(uint64(hostID), 10) + ":%"
}

// HostReferencedByZeroTrust 主机是否被零信任台账引用（供 host_service 删除前校验）
func HostReferencedByZeroTrust(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.ZeroTrust{}).
		Where(hostInTargetsCond, hostInTargetsPattern(hostID)).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询零信任台账引用失败: %w", err)
	}
	return count, nil
}

// ZeroTrustsByHost 查询引用指定主机的零信任台账记录（供主机详情展示）
func ZeroTrustsByHost(hostID uint) ([]models.ZeroTrust, error) {
	var records []models.ZeroTrust
	err := database.DB.
		Where(hostInTargetsCond, hostInTargetsPattern(hostID)).
		Order("id ASC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("查询零信任台账失败: %w", err)
	}
	return records, nil
}

// BatchCreateText 零信任批量添加：列顺序 申请单位,账户名,联系方式,内网IP,申请端口,系统名称,申请时间,备注[,公网IP]
// 内网IP与申请端口两列数量必须一致、按位置配对；公网IP第9列选填、须在资源池
// 主机按内网IP定位，不存在则该行失败；合法行全部插入（无唯一约束，不跳过）
func (s *ZeroTrustService) BatchCreateText(text string, op Operator, requestID string) (*BatchCreateResponse, error) {
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

		if !headerSkipped && len(row) > 0 && strings.TrimSpace(row[0]) == "申请单位" {
			headerSkipped = true
			continue
		}

		if len(row) > 9 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 列数超出 9 列", lineNo))
			continue
		}
		if len(row) < 4 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 至少需要4列(申请单位,账户名,联系方式,内网IP)", lineNo))
			continue
		}
		for len(row) < 9 {
			row = append(row, "")
		}

		ips, missing := lookupZeroTrustHostIDs(row[3])
		if len(missing) > 0 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网主机不存在(%s)", lineNo, strings.Join(missing, ",")))
			continue
		}
		if len(ips) == 0 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网IP不能为空", lineNo))
			continue
		}

		ports, err := parsePortList("申请端口", row[4])
		if err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
			continue
		}
		if len(ips) != len(ports) {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网IP与申请端口数量必须一致（IP %d 个，端口 %d 个）", lineNo, len(ips), len(ports)))
			continue
		}

		targets := make([]ZeroTrustTarget, 0, len(ips))
		for idx := range ips {
			p, convErr := strconv.Atoi(ports[idx])
			if convErr != nil || p < 1 || p > 65535 {
				resp.Errors++
				resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 申请端口必须在 1-65535 之间", lineNo))
				targets = nil
				break
			}
			targets = append(targets, ZeroTrustTarget{HostID: ips[idx], Port: p})
		}
		if targets == nil {
			continue
		}

		req := ZeroTrustRequest{
			ApplyUnit:   strings.TrimSpace(row[0]),
			AccountName: strings.TrimSpace(row[1]),
			Contact:     strings.TrimSpace(row[2]),
			PublicIP:    strings.TrimSpace(row[8]),
			Targets:     targets,
			SystemName:  strings.TrimSpace(row[5]),
			Remark:      strings.TrimSpace(row[7]),
		}
		if t := strings.TrimSpace(row[6]); t != "" {
			parsed, err := utils.ParseFlexibleTime(t)
			if err != nil {
				resp.Errors++
				resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 申请时间无效: %v", lineNo, err))
				continue
			}
			req.ApplyTime = &utils.FlexibleTime{Time: parsed}
		}

		if _, err := s.Create(req, op, requestID); err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
		} else {
			resp.Success++
		}
	}

	return resp, nil
}

// lookupZeroTrustHostIDs 按逗号分隔的内网IP列表查主机ID（保留顺序与重复，供位置配对），返回ID与缺失的IP
func lookupZeroTrustHostIDs(rawIPs string) ([]uint, []string) {
	ids := make([]uint, 0)
	missing := make([]string, 0)
	for _, ip := range strings.Split(rawIPs, ",") {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		var host models.Host
		if err := database.DB.Where("private_ip = ?", ip).First(&host).Error; err != nil {
			missing = append(missing, ip)
			continue
		}
		ids = append(ids, host.ID)
	}
	return ids, missing
}
