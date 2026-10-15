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

	"gorm.io/gorm"
)

type ZeroTrustService struct{}

func NewZeroTrustService() *ZeroTrustService {
	return &ZeroTrustService{}
}

// ZeroTrustTarget 一组接入目标配对；T-056 起端口为表达式（`22`/`1000-10005`/`22,443,5432` 可混合），
// 兼容旧 API 入参单端口数字字段 port（ports 为空时回退，响应仅返回 ports）
type ZeroTrustTarget struct {
	HostID uint   `json:"host_id"`
	Ports  string `json:"ports"`
	Port   *int   `json:"port,omitempty"`
}

type ZeroTrustRequest struct {
	ApplyUnit   string              `json:"apply_unit" binding:"required"`
	AccountName string              `json:"account_name" binding:"required"`
	Applicant   string              `json:"applicant"`
	Contact     string              `json:"contact"`
	PublicIP    string              `json:"public_ip"`
	Targets     []ZeroTrustTarget   `json:"targets" binding:"required"`
	SystemName  string              `json:"system_name"`
	ApplyTime   *utils.FlexibleTime `json:"apply_time"`
	Remark      string              `json:"remark"`
}

// ZeroTrustTargetView 配对 + 主机简要信息（列表展示用；ports 为端口表达式）
type ZeroTrustTargetView struct {
	HostID    uint   `json:"host_id"`
	Ports     string `json:"ports"`
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

// normalizeTargetPairs 校验配对列表：端口列表达式（复用 parsePortExpr，单口/范围规范化）、
// 兼容旧 port 数字入参、按「主机+表达式」去重，返回归一化配对
func normalizeTargetPairs(list []ZeroTrustTarget) ([]ZeroTrustTarget, error) {
	if len(list) == 0 {
		return nil, errors.New("接入主机与端口不能为空")
	}
	seen := make(map[string]struct{}, len(list))
	pairs := make([]ZeroTrustTarget, 0, len(list))
	for _, t := range list {
		if t.HostID == 0 {
			return nil, errors.New("申请主机不能为空")
		}
		raw := strings.TrimSpace(t.Ports)
		if raw == "" && t.Port != nil {
			raw = strconv.Itoa(*t.Port)
		}
		tokens, err := parsePortExpr("申请端口", raw)
		if err != nil {
			return nil, err
		}
		t.Ports = strings.Join(tokens, ",")
		t.Port = nil
		key := strconv.FormatUint(uint64(t.HostID), 10) + "|" + t.Ports
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		pairs = append(pairs, t)
	}
	return pairs, nil
}

// joinTargetPairs 配对列表 → "host_id:ports;host_id:ports" 串（派生字段/审计 diff 用；
// T-056 起端口为表达式，对间以 ; 分隔、对内表达式保留逗号，避免与旧逗号分隔歧义）
func joinTargetPairs(pairs []ZeroTrustTarget) string {
	parts := make([]string, len(pairs))
	for i, t := range pairs {
		parts[i] = strconv.FormatUint(uint64(t.HostID), 10) + ":" + t.Ports
	}
	return strings.Join(parts, ";")
}

// loadZeroTrustTargetPairs 批量加载接入目标子表（zero_trust_id → 配对，按行 id 升序）；
// db 可为事务（导入路径传 tx）
func loadZeroTrustTargetPairs(db *gorm.DB, ids []uint) (map[uint][]ZeroTrustTarget, error) {
	pairs := make(map[uint][]ZeroTrustTarget, len(ids))
	if len(ids) == 0 {
		return pairs, nil
	}
	var rows []models.ZeroTrustTarget
	if err := db.Where("zero_trust_id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询零信任接入目标失败: %w", err)
	}
	for _, r := range rows {
		pairs[r.ZeroTrustID] = append(pairs[r.ZeroTrustID], ZeroTrustTarget{HostID: r.HostID, Ports: r.Ports})
	}
	return pairs, nil
}

// insertZeroTrustTargets 事务内写入配对子表行（Create/Update/导入共用）
func insertZeroTrustTargets(tx *gorm.DB, zeroTrustID uint, pairs []ZeroTrustTarget) error {
	if len(pairs) == 0 {
		return nil
	}
	rows := make([]models.ZeroTrustTarget, 0, len(pairs))
	for _, p := range pairs {
		rows = append(rows, models.ZeroTrustTarget{ZeroTrustID: zeroTrustID, HostID: p.HostID, Ports: p.Ports})
	}
	return tx.Create(&rows).Error
}

// populateZeroTrustTargets 回填派生 Targets 字段（gorm:"-"，读时组装）
func populateZeroTrustTargets(records []models.ZeroTrust, pairs map[uint][]ZeroTrustTarget) {
	for i := range records {
		records[i].Targets = joinTargetPairs(pairs[records[i].ID])
	}
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
	req.Applicant = strings.TrimSpace(req.Applicant)
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
			"zero_trusts.apply_unit ILIKE ? OR zero_trusts.account_name ILIKE ? OR zero_trusts.applicant ILIKE ? OR zero_trusts.contact ILIKE ? OR zero_trusts.system_name ILIKE ? OR zero_trusts.remark ILIKE ? OR zero_trusts.public_ip ILIKE ? OR EXISTS (SELECT 1 FROM public_ips pi WHERE pi.ip = zero_trusts.public_ip AND (pi.exit_location ILIKE ? OR pi.isp ILIKE ?)) OR EXISTS (SELECT 1 FROM zero_trust_targets ztt JOIN hosts h ON h.id = ztt.host_id WHERE ztt.zero_trust_id = zero_trusts.id AND (h.name ILIKE ? OR h.private_ip ILIKE ?))",
			like, like, like, like, like, like, like, like, like, like, like,
		)
	}

	var records []models.ZeroTrust
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	ids := make([]uint, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	pairsByID, err := loadZeroTrustTargetPairs(database.DB, ids)
	if err != nil {
		return nil, err
	}

	hostIDSet := make(map[uint]struct{})
	for _, pairs := range pairsByID {
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
	for _, r := range records {
		pairs := pairsByID[r.ID]
		item := ZeroTrustItem{ZeroTrust: r, Targets: make([]ZeroTrustTargetView, 0, len(pairs))}
		for _, t := range pairs {
			view := ZeroTrustTargetView{HostID: t.HostID, Ports: t.Ports}
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
	pairs, err := loadZeroTrustTargetPairs(database.DB, []uint{record.ID})
	if err != nil {
		return nil, err
	}
	record.Targets = joinTargetPairs(pairs[record.ID])
	return &record, nil
}

// unionTargetPairs 两组配对并集去重（键=主机+端口表达式，保序：已有在前、新对追加）（T-058）
func unionTargetPairs(existing, incoming []ZeroTrustTarget) []ZeroTrustTarget {
	seen := make(map[string]struct{}, len(existing)+len(incoming))
	out := make([]ZeroTrustTarget, 0, len(existing)+len(incoming))
	for _, t := range existing {
		key := strconv.FormatUint(uint64(t.HostID), 10) + "|" + t.Ports
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	for _, t := range incoming {
		key := strconv.FormatUint(uint64(t.HostID), 10) + "|" + t.Ports
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}

// mergeZeroTrustInto 把一次新增合并进同账户名已有记录（T-058）：
// 接入目标取并集去重；其余字段已有非空保留、新值仅补空位（公网IP同规则且补录值须在资源池）；
// 申请时间保留已有；无实际变化时不写库
func mergeZeroTrustInto(dst *models.ZeroTrust, req ZeroTrustRequest, newPairs []ZeroTrustTarget, op Operator, requestID string) (uint, error) {
	curPairs, err := loadZeroTrustTargetPairs(database.DB, []uint{dst.ID})
	if err != nil {
		return 0, err
	}
	before := map[string]interface{}{
		"apply_unit":   dst.ApplyUnit,
		"account_name": dst.AccountName,
		"applicant":    dst.Applicant,
		"contact":      dst.Contact,
		"public_ip":    dst.PublicIP,
		"targets":      joinTargetPairs(curPairs[dst.ID]),
		"system_name":  dst.SystemName,
		"remark":       dst.Remark,
	}

	finalPublicIP := dst.PublicIP
	if finalPublicIP == "" {
		finalPublicIP = req.PublicIP
	}
	if err := validateZeroTrustPublicIP(finalPublicIP); err != nil {
		return 0, err
	}

	mergedPairs := unionTargetPairs(curPairs[dst.ID], newPairs)
	changed := false
	if dst.Applicant == "" && req.Applicant != "" {
		dst.Applicant = req.Applicant
		changed = true
	}
	if dst.Contact == "" && req.Contact != "" {
		dst.Contact = req.Contact
		changed = true
	}
	if dst.PublicIP == "" && req.PublicIP != "" {
		dst.PublicIP = req.PublicIP
		changed = true
	}
	if dst.SystemName == "" && req.SystemName != "" {
		dst.SystemName = req.SystemName
		changed = true
	}
	if dst.Remark == "" && req.Remark != "" {
		dst.Remark = req.Remark
		changed = true
	}
	if !changed && targetPairsEqual(curPairs[dst.ID], mergedPairs) {
		return dst.ID, nil
	}

	tx := database.DB.Begin()
	if err := tx.Save(dst).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Where("zero_trust_id = ?", dst.ID).Delete(&models.ZeroTrustTarget{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := insertZeroTrustTargets(tx, dst.ID, mergedPairs); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	after := map[string]interface{}{
		"apply_unit":   dst.ApplyUnit,
		"account_name": dst.AccountName,
		"applicant":    dst.Applicant,
		"contact":      dst.Contact,
		"public_ip":    dst.PublicIP,
		"targets":      joinTargetPairs(mergedPairs),
		"system_name":  dst.SystemName,
		"remark":       dst.Remark,
	}
	NewAuditService().Record(op, "update", "zero_trust", &dst.ID, DetailDiff{Before: before, After: after}, requestID)
	return dst.ID, nil
}

// Create 新增零信任；同账户名已有记录时合并进 id 最小的一条（T-058，存量同名多条不迁移）
func (s *ZeroTrustService) Create(req ZeroTrustRequest, op Operator, requestID string) (uint, error) {
	req, pairs, err := normalizeZeroTrust(req)
	if err != nil {
		return 0, err
	}
	if err := validateZeroTrustHosts(pairs); err != nil {
		return 0, err
	}

	var existing models.ZeroTrust
	findErr := database.DB.Where("account_name = ?", req.AccountName).Order("id ASC").First(&existing).Error
	if findErr == nil {
		return mergeZeroTrustInto(&existing, req, pairs, op, requestID)
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return 0, findErr
	}
	if err := validateZeroTrustPublicIP(req.PublicIP); err != nil {
		return 0, err
	}

	record := models.ZeroTrust{
		ApplyUnit:   req.ApplyUnit,
		AccountName: req.AccountName,
		Applicant:   req.Applicant,
		Contact:     req.Contact,
		PublicIP:    req.PublicIP,
		SystemName:  req.SystemName,
		ApplyTime:   req.ApplyTime.Time,
		Remark:      req.Remark,
	}
	tx := database.DB.Begin()
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := insertZeroTrustTargets(tx, record.ID, pairs); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit().Error; err != nil {
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
	curPairs, pairErr := loadZeroTrustTargetPairs(database.DB, []uint{record.ID})
	if pairErr != nil {
		return pairErr
	}

	before := map[string]interface{}{
		"apply_unit":   record.ApplyUnit,
		"account_name": record.AccountName,
		"applicant":    record.Applicant,
		"contact":      record.Contact,
		"public_ip":    record.PublicIP,
		"targets":      joinTargetPairs(curPairs[record.ID]),
		"system_name":  record.SystemName,
		"remark":       record.Remark,
	}
	record.ApplyUnit = req.ApplyUnit
	record.AccountName = req.AccountName
	record.Applicant = req.Applicant
	record.Contact = req.Contact
	record.PublicIP = req.PublicIP
	record.SystemName = req.SystemName
	record.ApplyTime = req.ApplyTime.Time
	record.Remark = req.Remark
	tx := database.DB.Begin()
	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("zero_trust_id = ?", record.ID).Delete(&models.ZeroTrustTarget{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := insertZeroTrustTargets(tx, record.ID, pairs); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
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
	curPairs, err := loadZeroTrustTargetPairs(database.DB, []uint{record.ID})
	if err != nil {
		return err
	}
	tx := database.DB.Begin()
	if err := tx.Where("zero_trust_id = ?", record.ID).Delete(&models.ZeroTrustTarget{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "delete", "zero_trust", &id,
		DetailDiff{Before: map[string]interface{}{
			"apply_unit": record.ApplyUnit, "account_name": record.AccountName, "applicant": record.Applicant,
			"contact": record.Contact, "public_ip": record.PublicIP, "targets": joinTargetPairs(curPairs[record.ID]),
			"system_name": record.SystemName, "remark": record.Remark,
		}}, requestID)
	return nil
}

// HostReferencedByZeroTrust 引用指定主机的零信任台账记录数（供 host_service 删除前校验；
// 子表按 zero_trust_id 去重计数，同记录多配对只计一次）
func HostReferencedByZeroTrust(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.ZeroTrustTarget{}).
		Distinct("zero_trust_id").
		Where("host_id = ?", hostID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询零信任台账引用失败: %w", err)
	}
	return count, nil
}

// ZeroTrustsByHost 查询引用指定主机的零信任台账记录（供主机详情展示，派生 Targets 已回填）
func ZeroTrustsByHost(hostID uint) ([]models.ZeroTrust, error) {
	var records []models.ZeroTrust
	err := database.DB.
		Where("id IN (SELECT zero_trust_id FROM zero_trust_targets WHERE host_id = ?)", hostID).
		Order("id ASC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("查询零信任台账失败: %w", err)
	}
	ids := make([]uint, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	pairs, err := loadZeroTrustTargetPairs(database.DB, ids)
	if err != nil {
		return nil, err
	}
	populateZeroTrustTargets(records, pairs)
	return records, nil
}

// BatchCreateText 零信任批量添加：列顺序 申请单位,账户名,申请人,联系方式,内网IP,申请端口,系统名称,申请时间,备注[,公网IP]
// （T-060 起第3列新增申请人，联系方式后移至第4列；旧 9 列文件因申请人列缺失会被视为联系方式→需注意，故按 10 列口径）
// 内网IP与申请端口两列数量必须一致、按位置配对；申请端口列含 `;` 时按 `;` 分组（T-058，多IP各自多端口），
// 无 `;` 且仅 1 个端口组时广播到全部 IP（T-059，多IP相同端口）；单IP+整列逗号端口视为该主机多端口；
// 公网IP第10列选填、须在资源池；主机按内网IP定位，不存在则该行失败；
// 合法行全部插入；同账户名行合并进已有记录（T-058）
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

		if len(row) > 10 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 列数超出 10 列", lineNo))
			continue
		}
		if len(row) < 5 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 至少需要5列(申请单位,账户名,申请人,联系方式,内网IP)", lineNo))
			continue
		}
		for len(row) < 10 {
			row = append(row, "")
		}

		ips, missing := lookupZeroTrustHostIDs(row[4])
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

		groups, err := splitPortGroups(row[5])
		if err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
			continue
		}
		// T-058：单IP+整列逗号端口（无 `;` 分组）直接视为该主机多端口表达式，无需双层引号
		if len(ips) == 1 && len(groups) > 1 && !strings.Contains(row[5], ";") {
			groups = []string{strings.Join(groups, ",")}
		}
		// T-059：无 `;` 且仅 1 个端口组 → 广播到全部 IP（多IP相同端口；此前报数不一致）
		if len(groups) == 1 && len(ips) > 1 && !strings.Contains(row[5], ";") {
			broadcast := make([]string, len(ips))
			for i := range ips {
				broadcast[i] = groups[0]
			}
			groups = broadcast
		}
		if len(ips) != len(groups) {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网IP与申请端口数量必须一致（IP %d 个，端口组 %d 个）", lineNo, len(ips), len(groups)))
			continue
		}

		targets := make([]ZeroTrustTarget, 0, len(ips))
		for idx := range ips {
			targets = append(targets, ZeroTrustTarget{HostID: ips[idx], Ports: groups[idx]})
		}

		req := ZeroTrustRequest{
			ApplyUnit:   strings.TrimSpace(row[0]),
			AccountName: strings.TrimSpace(row[1]),
			Applicant:   strings.TrimSpace(row[2]),
			Contact:     strings.TrimSpace(row[3]),
			PublicIP:    strings.TrimSpace(row[9]),
			Targets:     targets,
			SystemName:  strings.TrimSpace(row[6]),
			Remark:      strings.TrimSpace(row[8]),
		}
		if t := strings.TrimSpace(row[7]); t != "" {
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

// splitPortGroups 解析批量「申请端口」列（T-056/T-058）：含 `;` 时按 `;` 分组（T-058，
// 组内逗号=同主机多端口/范围，多IP各自不同端口）；无 `;` 时顶层逗号分组，
// 双引号包裹的组允许含逗号（旧双层引号路径，如 "22,443",8080），未加引号时每 token 一组（旧文件平铺兼容）
func splitPortGroups(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("申请端口不能为空")
	}
	if strings.Contains(raw, ";") {
		var groups []string
		for _, seg := range strings.Split(raw, ";") {
			if s := strings.TrimSpace(seg); s != "" {
				groups = append(groups, s)
			}
		}
		if len(groups) == 0 {
			return nil, errors.New("申请端口不能为空")
		}
		return groups, nil
	}
	var groups []string
	var cur strings.Builder
	inQuote := false
	for _, r := range raw {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ',' && !inQuote:
			if s := strings.TrimSpace(cur.String()); s != "" {
				groups = append(groups, s)
			}
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if inQuote {
		return nil, errors.New("申请端口引号未闭合")
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		groups = append(groups, s)
	}
	if len(groups) == 0 {
		return nil, errors.New("申请端口不能为空")
	}
	return groups, nil
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
