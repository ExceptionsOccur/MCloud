package services

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// errImportRolledBack 行级校验失败哨兵：触发事务回滚（整体不落库）
var errImportRolledBack = errors.New("导入校验失败，已整体回滚")

// ErrImportStructure 结构性错误哨兵：缺 sheet/缺列/表头重复/非 xlsx（controller 映射 40001）
var ErrImportStructure = errors.New("文件结构错误")

// importSheet 单个 sheet 的解析结果（表头 → 列下标 + 数据行）
type importSheet struct {
	name   string
	colIdx map[string]int
	rows   [][]string
}

// get 按表头取单元格值（列缺失/越界返回空串）
func (sh *importSheet) get(row []string, header string) string {
	idx, ok := sh.colIdx[header]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// isEmpty 整行所有定义列均为空视为空行（不计入数据行）
func (sh *importSheet) isEmpty(row []string) bool {
	for _, idx := range sh.colIdx {
		if idx < len(row) && strings.TrimSpace(row[idx]) != "" {
			return false
		}
	}
	return true
}

// parseWorkbook 结构校验：8 个 sheet 齐全且表头与定义一致（按名匹配列，位置无关）
func parseWorkbook(data []byte) (map[string]*importSheet, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("文件解析失败，请上传 .xlsx 文件")
	}
	defer f.Close()

	sheets := make(map[string]*importSheet, len(dataExchangeSheets))
	for _, spec := range dataExchangeSheets {
		rows, err := f.GetRows(spec.Name)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				return nil, fmt.Errorf("缺少工作表 %q", spec.Name)
			}
			return nil, fmt.Errorf("读取工作表 %q 失败: %v", spec.Name, err)
		}
		if len(rows) == 0 {
			return nil, fmt.Errorf("缺少工作表 %q", spec.Name)
		}
		colIdx := make(map[string]int, len(rows[0]))
		for i, h := range rows[0] {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			if _, dup := colIdx[h]; dup {
				return nil, fmt.Errorf("工作表 %q 表头重复: %q", spec.Name, h)
			}
			colIdx[h] = i
		}
		var missing []string
		for _, need := range spec.Headers {
			if _, ok := colIdx[need]; !ok {
				missing = append(missing, need)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("工作表 %q 缺少列: %s", spec.Name, strings.Join(missing, ", "))
		}
		sheets[spec.Name] = &importSheet{name: spec.Name, colIdx: colIdx, rows: rows[1:]}
	}
	return sheets, nil
}

// Import 导入 xlsx 并按依赖序 upsert 8 表；单事务，任意行失败整体回滚
func (s *DataExchangeService) Import(data []byte, op Operator, requestID string) (*ImportReport, error) {
	sheets, err := parseWorkbook(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrImportStructure, err.Error())
	}

	report := newImportReport()
	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		steps := []func(*gorm.DB, *importSheet, *ImportReport) error{
			s.applyPersons,
			s.applyPublicIPs,
			s.applyCloudResources,
			s.applySubnets,
			s.applyHosts,
			s.applyHostApplications,
			s.applyZeroTrusts,
			s.applyPortMappings,
		}
		for i, step := range steps {
			if err := step(tx, sheets[dataExchangeSheets[i].Name], report); err != nil {
				return err
			}
		}
		if len(report.Errors) > 0 {
			return errImportRolledBack
		}
		// ip_mapped 为派生列：按映射表全量重算
		return tx.Exec("UPDATE hosts SET ip_mapped = EXISTS (SELECT 1 FROM port_mappings pm WHERE pm.host_id = hosts.id)").Error
	})
	if txErr != nil {
		if errors.Is(txErr, errImportRolledBack) {
			report.Committed = false
			return report, nil
		}
		return nil, fmt.Errorf("导入执行失败: %w", txErr)
	}
	report.Committed = true
	NewAuditService().Record(op, "import_xlsx", "data_exchange", nil,
		DetailDiff{After: map[string]interface{}{"committed": true}}, requestID)
	return report, nil
}

// ---------- 通用工具 ----------

func newImportReport() *ImportReport {
	r := &ImportReport{Errors: []RowError{}}
	for _, spec := range dataExchangeSheets {
		r.Sheets = append(r.Sheets, SheetStat{Sheet: spec.Name})
	}
	return r
}

// stat 定位 sheet 统计（预填充，恒命中）
func (r *ImportReport) stat(sheet string) *SheetStat {
	for i := range r.Sheets {
		if r.Sheets[i].Sheet == sheet {
			return &r.Sheets[i]
		}
	}
	return nil
}

// addErr 记录行级错误（row = Excel 实际行号，表头为第 1 行）
func (r *ImportReport) addErr(sheet string, row int, msg string) {
	r.Errors = append(r.Errors, RowError{Sheet: sheet, Row: row, Message: msg})
}

// parseIntCell 空 → 0；整数/浮点整值 → int；否则报错
func parseIntCell(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	if n, err := strconv.Atoi(raw); err == nil {
		return n, nil
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil && f == float64(int(f)) {
		return int(f), nil
	}
	return 0, fmt.Errorf("数字格式不正确 %q", raw)
}

// parseBoolCell 是/否/true/false/1/0；空 → false；其余报错
func parseBoolCell(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "否", "false", "0", "no":
		return false, nil
	case "是", "true", "1", "yes":
		return true, nil
	}
	return false, fmt.Errorf("布尔值格式不正确 %q（应为 是/否）", raw)
}

// parseTimeCell 依次尝试常见布局（无时区按服务器本地时区解释），空 → zero
func parseTimeCell(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("时间格式不正确 %q（示例: 2026-10-09 15:04:05）", raw)
}

// rowCtx 单数据行上下文
type rowCtx struct {
	sheet  string
	rowNum int
	sh     *importSheet
	row    []string
}

func (c *rowCtx) get(header string) string { return c.sh.get(c.row, header) }

// fail 记录行级错误（cb 内提前返回用）
func fail(report *ImportReport, c *rowCtx, err error) {
	report.addErr(c.sheet, c.rowNum, err.Error())
}

// iterRows 遍历非空数据行并累计行数；cb 内自行 fail + return
func iterRows(sh *importSheet, report *ImportReport, cb func(c *rowCtx, stat *SheetStat)) {
	stat := report.stat(sh.name)
	for i, row := range sh.rows {
		if sh.isEmpty(row) {
			continue
		}
		stat.Rows++
		cb(&rowCtx{sheet: sh.name, rowNum: i + 2, sh: sh, row: row}, stat)
	}
}

// resolvePerson 按姓名(+联系方式)解析人员；空姓名 → nil；未找到 → error
func resolvePerson(tx *gorm.DB, name, contact string) (*uint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	contact = strings.TrimSpace(contact)
	var p models.Person
	var err error
	if contact != "" {
		err = tx.Where("name = ? AND contact = ?", name, contact).Order("id ASC").First(&p).Error
	} else {
		err = tx.Where("name = ?", name).Order("id ASC").First(&p).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("人员不存在: %s", name)
	}
	if err != nil {
		return nil, err
	}
	return &p.ID, nil
}

// resolveHostIP 按内网IP解析主机ID
func resolveHostIP(tx *gorm.DB, ip string) (uint, error) {
	var h models.Host
	err := tx.Where("private_ip = ?", ip).Order("id ASC").First(&h).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("主机不存在: %s", ip)
	}
	if err != nil {
		return 0, err
	}
	return h.ID, nil
}

// requirePublicIP 公网IP非空时必须在资源池中
func requirePublicIP(tx *gorm.DB, ip string) error {
	if strings.TrimSpace(ip) == "" {
		return nil
	}
	var count int64
	if err := tx.Model(&models.PublicIP{}).Where("ip = ?", ip).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrPublicIPNotInPool
	}
	return nil
}

func uintPtrEqual(a, b *uint) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func personKey(name, contact string) string {
	return strings.TrimSpace(name) + "\x00" + strings.TrimSpace(contact)
}

// ---------- 各 sheet 导入（自然键 upsert：存在且一致→跳过，存在不一致→更新，不存在→新增） ----------

// applyPersons 自然键 name+contact
func (s *DataExchangeService) applyPersons(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.Person
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.Person, len(existing))
	for i := range existing {
		byKey[personKey(existing[i].Name, existing[i].Contact)] = &existing[i]
	}
	seen := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		name := c.get("姓名")
		if name == "" {
			fail(report, c, errors.New("姓名不能为空"))
			return
		}
		contact := c.get("联系方式")
		unit := c.get("单位")
		key := personKey(name, contact)
		if first, dup := seen[key]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行人员重复（姓名+联系方式）", first))
			return
		}
		seen[key] = c.rowNum

		if cur, ok := byKey[key]; ok {
			if cur.Unit == unit {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.Person{}).Where("id = ?", cur.ID).
				Update("unit", unit).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		p := models.Person{Name: name, Contact: contact, Unit: unit}
		if err := tx.Create(&p).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[key] = &p
		stat.Created++
	})
	return nil
}

// applyPublicIPs 自然键 ip
func (s *DataExchangeService) applyPublicIPs(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.PublicIP
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.PublicIP, len(existing))
	for i := range existing {
		byKey[existing[i].IP] = &existing[i]
	}
	seen := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		ip := c.get("公网IP")
		if ip == "" {
			fail(report, c, errors.New("公网IP不能为空"))
			return
		}
		if net.ParseIP(ip) == nil {
			fail(report, c, fmt.Errorf("公网IP格式不正确 %q", ip))
			return
		}
		if first, dup := seen[ip]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行公网IP重复", first))
			return
		}
		seen[ip] = c.rowNum

		isp, exitLoc, remark := c.get("运营商"), c.get("出口位置"), c.get("备注")
		if cur, ok := byKey[ip]; ok {
			if cur.ISP == isp && cur.ExitLocation == exitLoc && cur.Remark == remark {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.PublicIP{}).Where("id = ?", cur.ID).Updates(map[string]interface{}{
				"isp": isp, "exit_location": exitLoc, "remark": remark,
			}).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		r := models.PublicIP{IP: ip, ISP: isp, ExitLocation: exitLoc, Remark: remark}
		if err := tx.Create(&r).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[ip] = &r
		stat.Created++
	})
	return nil
}

// applyCloudResources 自然键 region
func (s *DataExchangeService) applyCloudResources(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.CloudResource
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.CloudResource, len(existing))
	for i := range existing {
		byKey[existing[i].Region] = &existing[i]
	}
	seen := make(map[string]int)

	intCols := []string{"物理CPU", "VCPU", "内存", "存储", "裸金属", "GPU卡数", "对象存储"}
	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		region := c.get("区域")
		if region == "" {
			fail(report, c, errors.New("区域不能为空"))
			return
		}
		if first, dup := seen[region]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行区域重复", first))
			return
		}
		seen[region] = c.rowNum

		vals := make([]int, len(intCols))
		for i, col := range intCols {
			n, err := parseIntCell(c.get(col))
			if err != nil {
				fail(report, c, fmt.Errorf("%s: %v", col, err))
				return
			}
			vals[i] = n
		}
		updates := map[string]interface{}{
			"physical_cpu": vals[0], "vcpu": vals[1], "memory": vals[2], "storage": vals[3],
			"bare_metal": vals[4], "gpu_card_count": vals[5], "object_storage": vals[6],
		}
		if cur, ok := byKey[region]; ok {
			if cur.PhysicalCPU == vals[0] && cur.Vcpu == vals[1] && cur.Memory == vals[2] &&
				cur.Storage == vals[3] && cur.BareMetal == vals[4] &&
				cur.GpuCardCount == vals[5] && cur.ObjectStorage == vals[6] {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.CloudResource{}).Where("id = ?", cur.ID).Updates(updates).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		r := models.CloudResource{
			Region: region, PhysicalCPU: vals[0], Vcpu: vals[1], Memory: vals[2],
			Storage: vals[3], BareMetal: vals[4], GpuCardCount: vals[5], ObjectStorage: vals[6],
		}
		if err := tx.Create(&r).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[region] = &r
		stat.Created++
	})
	return nil
}

// applySubnets 自然键 cidr（复用 /24 规范化校验）
func (s *DataExchangeService) applySubnets(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.IPSubnet
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.IPSubnet, len(existing))
	for i := range existing {
		byKey[existing[i].CIDR] = &existing[i]
	}
	seen := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		cidr, err := normalizeCIDR(c.get("网段CIDR"))
		if err != nil {
			fail(report, c, err)
			return
		}
		if first, dup := seen[cidr]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行网段重复", first))
			return
		}
		seen[cidr] = c.rowNum
		if _, ok := byKey[cidr]; ok {
			stat.Skipped++
			return
		}
		sub := models.IPSubnet{CIDR: cidr}
		if err := tx.Create(&sub).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[cidr] = &sub
		stat.Created++
	})
	return nil
}

// applyHosts 自然键 private_ip；人员列按姓名(+联系方式)解析 person_id
func (s *DataExchangeService) applyHosts(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.Host
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.Host, len(existing))
	for i := range existing {
		byKey[existing[i].PrivateIP] = &existing[i]
	}
	seen := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		region, name := c.get("区域"), c.get("主机名称")
		privateIP := c.get("内网IP")
		if region == "" {
			fail(report, c, errors.New("区域不能为空"))
			return
		}
		if name == "" {
			fail(report, c, errors.New("主机名称不能为空"))
			return
		}
		if privateIP == "" {
			fail(report, c, errors.New("内网IP不能为空"))
			return
		}
		if first, dup := seen[privateIP]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行内网IP重复", first))
			return
		}
		seen[privateIP] = c.rowNum

		cpu, err := parseIntCell(c.get("CPU核数"))
		if err != nil {
			fail(report, c, fmt.Errorf("CPU核数: %v", err))
			return
		}
		memory, err := parseIntCell(c.get("内存"))
		if err != nil {
			fail(report, c, fmt.Errorf("内存: %v", err))
			return
		}
		systemDisk, err := parseIntCell(c.get("系统盘"))
		if err != nil {
			fail(report, c, fmt.Errorf("系统盘: %v", err))
			return
		}
		dataDisk, err := parseIntCell(c.get("数据盘"))
		if err != nil {
			fail(report, c, fmt.Errorf("数据盘: %v", err))
			return
		}
		isDB, err := parseBoolCell(c.get("是否数据库服务器"))
		if err != nil {
			fail(report, c, err)
			return
		}
		personID, err := resolvePerson(tx, c.get("人员姓名"), c.get("人员联系方式"))
		if err != nil {
			fail(report, c, err)
			return
		}

		f := struct {
			region, instanceID, assetType, os, cpuArch, envType, status, openPorts, tags string
		}{
			region: region,
			//nolint:revive // 字段与列同名，保持与导出列一致
			instanceID: c.get("实例ID"),
			assetType:  c.get("资产类型"),
			os:         c.get("操作系统"),
			cpuArch:    c.get("CPU架构"),
			envType:    c.get("环境类型"),
			status:     c.get("状态"),
			openPorts:  c.get("开放端口"),
			tags:       c.get("标签"),
		}
		updates := map[string]interface{}{
			"region": region, "instance_id": f.instanceID, "name": name,
			"asset_type": f.assetType, "os": f.os, "cpu": cpu, "cpu_arch": f.cpuArch,
			"memory": memory, "system_disk": systemDisk, "data_disk": dataDisk,
			"env_type": f.envType, "is_db_server": isDB, "status": f.status,
			"open_ports": f.openPorts, "tags": f.tags, "person_id": personID,
		}
		if cur, ok := byKey[privateIP]; ok {
			if cur.Region == f.region && cur.InstanceID == f.instanceID && cur.Name == name &&
				cur.AssetType == f.assetType && cur.OS == f.os && cur.CPU == cpu &&
				cur.CPUArch == f.cpuArch && cur.Memory == memory &&
				cur.SystemDisk == systemDisk && cur.DataDisk == dataDisk &&
				cur.EnvType == f.envType && cur.IsDBServer == isDB &&
				cur.Status == f.status && cur.OpenPorts == f.openPorts && cur.Tags == f.tags &&
				uintPtrEqual(cur.PersonID, personID) {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.Host{}).Where("id = ?", cur.ID).Updates(updates).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		h := models.Host{
			Region: region, InstanceID: f.instanceID, Name: name, PrivateIP: privateIP,
			AssetType: f.assetType, OS: f.os, CPU: cpu, CPUArch: f.cpuArch, Memory: memory,
			SystemDisk: systemDisk, DataDisk: dataDisk, EnvType: f.envType,
			IsDBServer: isDB, Status: f.status, OpenPorts: f.openPorts, Tags: f.tags,
			PersonID: personID,
		}
		if err := tx.Create(&h).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[privateIP] = &h
		stat.Created++
	})
	return nil
}

// applyHostApplications 自然键 = 宿主内网IP（host_id 唯一索引）
func (s *DataExchangeService) applyHostApplications(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.HostApplication
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byHost := make(map[uint]*models.HostApplication, len(existing))
	for i := range existing {
		byHost[existing[i].HostID] = &existing[i]
	}
	seen := make(map[string]int)

	strCols := []string{"申请单位", "申请人", "申请人联系方式", "所属项目", "申请理由", "申请配置", "申请时间", "对象存储大小", "备注"}
	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		privateIP := c.get("内网IP")
		if privateIP == "" {
			fail(report, c, errors.New("内网IP不能为空"))
			return
		}
		if first, dup := seen[privateIP]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行内网IP重复", first))
			return
		}
		seen[privateIP] = c.rowNum

		hostID, err := resolveHostIP(tx, privateIP)
		if err != nil {
			fail(report, c, err)
			return
		}
		vals := make([]string, len(strCols))
		for i, col := range strCols {
			vals[i] = c.get(col)
		}
		vals[6] = strings.TrimSpace(vals[6])
		if err := utils.ValidateDateOnly("申请时间", vals[6]); err != nil {
			fail(report, c, err)
			return
		}
		updates := map[string]interface{}{
			"apply_unit": vals[0], "applicant": vals[1], "applicant_contact": vals[2],
			"project": vals[3], "apply_reason": vals[4], "apply_config": vals[5],
			"apply_time": vals[6], "object_storage_size": vals[7], "remark": vals[8],
		}
		if cur, ok := byHost[hostID]; ok {
			if cur.ApplyUnit == vals[0] && cur.Applicant == vals[1] && cur.ApplicantContact == vals[2] &&
				cur.Project == vals[3] && cur.ApplyReason == vals[4] && cur.ApplyConfig == vals[5] &&
				cur.ApplyTime == vals[6] && cur.ObjectStorageSize == vals[7] && cur.Remark == vals[8] {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.HostApplication{}).Where("id = ?", cur.ID).Updates(updates).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		a := models.HostApplication{
			HostID: hostID, ApplyUnit: vals[0], Applicant: vals[1], ApplicantContact: vals[2],
			Project: vals[3], ApplyReason: vals[4], ApplyConfig: vals[5],
			ApplyTime: vals[6], ObjectStorageSize: vals[7], Remark: vals[8],
		}
		if err := tx.Create(&a).Error; err != nil {
			fail(report, c, err)
			return
		}
		byHost[hostID] = &a
		stat.Created++
	})
	return nil
}

// applyZeroTrusts 自然键 apply_unit+account_name+system_name；
// 接入目标 = 内网IP:端口（导入解析回 host_id），公网IP须在资源池
func (s *DataExchangeService) applyZeroTrusts(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.ZeroTrust
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.ZeroTrust, len(existing))
	for i := range existing {
		byKey[zeroTrustKey(existing[i])] = &existing[i]
	}
	seen := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		applyUnit := c.get("申请单位")
		accountName := c.get("账户名")
		systemName := c.get("系统名称")
		if applyUnit == "" {
			fail(report, c, errors.New("申请单位不能为空"))
			return
		}
		if accountName == "" {
			fail(report, c, errors.New("账户名不能为空"))
			return
		}
		key := strings.Join([]string{applyUnit, accountName, systemName}, "\x00")
		if first, dup := seen[key]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行记录重复（申请单位+账户名+系统名称）", first))
			return
		}
		seen[key] = c.rowNum

		contact := c.get("联系方式")
		publicIP := c.get("公网IP")
		remark := c.get("备注")

		targetsRaw := c.get("接入目标")
		if targetsRaw == "" {
			fail(report, c, errors.New("接入目标不能为空"))
			return
		}
		parts := make([]ZeroTrustTarget, 0)
		for _, seg := range strings.Split(targetsRaw, ",") {
			seg = strings.TrimSpace(seg)
			if seg == "" {
				continue
			}
			ip, portStr, ok := strings.Cut(seg, ":")
			if !ok {
				fail(report, c, fmt.Errorf("接入目标格式错误 %q（应为 内网IP:端口）", seg))
				return
			}
			port, err := strconv.Atoi(strings.TrimSpace(portStr))
			if err != nil || port < 1 || port > 65535 {
				fail(report, c, fmt.Errorf("接入目标端口不合法 %q（范围 1-65535）", seg))
				return
			}
			hostID, err := resolveHostIP(tx, strings.TrimSpace(ip))
			if err != nil {
				fail(report, c, err)
				return
			}
			parts = append(parts, ZeroTrustTarget{HostID: hostID, Port: port})
		}
		normalized, err := normalizeTargetPairs(parts)
		if err != nil {
			fail(report, c, err)
			return
		}
		if err = requirePublicIP(tx, publicIP); err != nil {
			fail(report, c, err)
			return
		}
		applyTime, err := parseTimeCell(c.get("申请时间"))
		if err != nil {
			fail(report, c, err)
			return
		}

		targets := joinTargetPairs(normalized)
		updates := map[string]interface{}{
			"contact": contact, "public_ip": publicIP, "targets": targets, "remark": remark,
		}
		if !applyTime.IsZero() {
			updates["apply_time"] = applyTime
		}

		if cur, ok := byKey[key]; ok {
			changed := cur.Contact != contact || cur.PublicIP != publicIP ||
				cur.Targets != targets || cur.Remark != remark ||
				(!applyTime.IsZero() && !cur.ApplyTime.Equal(applyTime))
			if !changed {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.ZeroTrust{}).Where("id = ?", cur.ID).Updates(updates).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		z := models.ZeroTrust{
			ApplyUnit: applyUnit, AccountName: accountName, SystemName: systemName,
			Contact: contact, PublicIP: publicIP, Targets: targets, Remark: remark,
			ApplyTime: time.Now(),
		}
		if !applyTime.IsZero() {
			z.ApplyTime = applyTime
		}
		if err := tx.Create(&z).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[key] = &z
		stat.Created++
	})
	return nil
}

func zeroTrustKey(z models.ZeroTrust) string {
	return strings.Join([]string{z.ApplyUnit, z.AccountName, z.SystemName}, "\x00")
}

// applyPortMappings 自然键 public_ip+宿主ID+外网端口；公网IP须在资源池；
// 域名非空须唯一（文件内 + 库内异键冲突）
func (s *DataExchangeService) applyPortMappings(tx *gorm.DB, sh *importSheet, report *ImportReport) error {
	var existing []models.PortMapping
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}
	byKey := make(map[string]*models.PortMapping, len(existing))
	byDomain := make(map[string]*models.PortMapping)
	for i := range existing {
		byKey[portMappingKey(existing[i].PublicIP, existing[i].HostID, existing[i].ExternalPorts)] = &existing[i]
		if existing[i].Domain != "" {
			byDomain[existing[i].Domain] = &existing[i]
		}
	}
	seen := make(map[string]int)
	seenDomain := make(map[string]int)

	iterRows(sh, report, func(c *rowCtx, stat *SheetStat) {
		publicIP := c.get("公网IP")
		hostIP := c.get("内网IP")
		domain := c.get("域名")
		remark := c.get("备注")
		if publicIP == "" {
			fail(report, c, errors.New("公网IP不能为空"))
			return
		}
		if hostIP == "" {
			fail(report, c, errors.New("内网IP不能为空"))
			return
		}
		extPorts, err := parsePortList("外网端口", c.get("外网端口"))
		if err != nil {
			fail(report, c, err)
			return
		}
		intlPorts, err := parsePortList("内网端口", c.get("内网端口"))
		if err != nil {
			fail(report, c, err)
			return
		}
		if len(extPorts) != len(intlPorts) {
			fail(report, c, fmt.Errorf("外网端口与内网端口数量必须一致（外网 %d 个，内网 %d 个）", len(extPorts), len(intlPorts)))
			return
		}
		extJoined := strings.Join(extPorts, ",")
		intlJoined := strings.Join(intlPorts, ",")

		hostID, err := resolveHostIP(tx, hostIP)
		if err != nil {
			fail(report, c, err)
			return
		}
		if err := requirePublicIP(tx, publicIP); err != nil {
			fail(report, c, err)
			return
		}
		key := portMappingKey(publicIP, hostID, extJoined)
		if first, dup := seen[key]; dup {
			fail(report, c, fmt.Errorf("与第 %d 行映射重复（公网IP+内网IP+外网端口）", first))
			return
		}
		seen[key] = c.rowNum
		if domain != "" {
			if first, dup := seenDomain[domain]; dup {
				fail(report, c, fmt.Errorf("与第 %d 行域名重复", first))
				return
			}
			seenDomain[domain] = c.rowNum
			if cur, dup := byDomain[domain]; dup &&
				portMappingKey(cur.PublicIP, cur.HostID, cur.ExternalPorts) != key {
				fail(report, c, fmt.Errorf("域名 %q 已被其他映射记录占用", domain))
				return
			}
		}

		updates := map[string]interface{}{
			"internal_ports": intlJoined, "domain": domain, "remark": remark,
		}
		if cur, ok := byKey[key]; ok {
			if cur.InternalPorts == intlJoined && cur.Domain == domain && cur.Remark == remark {
				stat.Skipped++
				return
			}
			if err := tx.Model(&models.PortMapping{}).Where("id = ?", cur.ID).Updates(updates).Error; err != nil {
				fail(report, c, err)
				return
			}
			stat.Updated++
			return
		}
		m := models.PortMapping{
			PublicIP: publicIP, HostID: hostID,
			ExternalPorts: extJoined, InternalPorts: intlJoined,
			Domain: domain, Remark: remark,
		}
		if err := tx.Create(&m).Error; err != nil {
			fail(report, c, err)
			return
		}
		byKey[key] = &m
		if domain != "" {
			byDomain[domain] = &m
		}
		stat.Created++
	})
	return nil
}

func portMappingKey(publicIP string, hostID uint, externalPorts string) string {
	return publicIP + "\x00" + strconv.FormatUint(uint64(hostID), 10) + "\x00" + externalPorts
}
