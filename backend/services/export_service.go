package services

import (
	"fmt"
	"time"

	"mcloud/database"
	"mcloud/models"

	"github.com/xuri/excelize/v2"
)

// Export 导出 8 张业务表为单个 xlsx 工作簿。
// sheet 顺序与 dataExchangeSheets 一致（依赖序）；关联一律用自然键表达：
// hosts 人员姓名/联系方式（非 person_id）、host_applications 内网IP（非 host_id）、
// zero_trusts 接入目标 = 内网IP:端口（非 host_id:port）、port_mappings 内网IP（非 host_id）。
func (s *DataExchangeService) Export() ([]byte, error) {
	var persons []models.Person
	if err := database.DB.Order("id ASC").Find(&persons).Error; err != nil {
		return nil, fmt.Errorf("查询人员失败: %w", err)
	}
	var publicIPs []models.PublicIP
	if err := database.DB.Order("id ASC").Find(&publicIPs).Error; err != nil {
		return nil, fmt.Errorf("查询公网IP失败: %w", err)
	}
	var cloudResources []models.CloudResource
	if err := database.DB.Order("region ASC").Find(&cloudResources).Error; err != nil {
		return nil, fmt.Errorf("查询云资源失败: %w", err)
	}
	var subnets []models.IPSubnet
	if err := database.DB.Order("cidr ASC").Find(&subnets).Error; err != nil {
		return nil, fmt.Errorf("查询IP网段失败: %w", err)
	}
	var hosts []models.Host
	if err := database.DB.Preload("Person").Preload("Application").Order("id ASC").Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("查询主机失败: %w", err)
	}
	var zeroTrusts []models.ZeroTrust
	if err := database.DB.Order("id ASC").Find(&zeroTrusts).Error; err != nil {
		return nil, fmt.Errorf("查询零信任台账失败: %w", err)
	}
	ztIDs := make([]uint, len(zeroTrusts))
	for i, z := range zeroTrusts {
		ztIDs[i] = z.ID
	}
	ztPairs, pairErr := loadZeroTrustTargetPairs(database.DB, ztIDs)
	if pairErr != nil {
		return nil, pairErr
	}
	var mappings []models.PortMapping
	if err := database.DB.Preload("Host").Order("id ASC").Find(&mappings).Error; err != nil {
		return nil, fmt.Errorf("查询端口映射失败: %w", err)
	}

	hostIPByID := make(map[uint]string, len(hosts))
	for _, h := range hosts {
		hostIPByID[h.ID] = h.PrivateIP
	}

	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", dataExchangeSheets[0].Name); err != nil {
		return nil, fmt.Errorf("创建工作表失败: %w", err)
	}
	for _, spec := range dataExchangeSheets[1:] {
		if _, err := f.NewSheet(spec.Name); err != nil {
			return nil, fmt.Errorf("创建工作表 %q 失败: %w", spec.Name, err)
		}
	}

	// persons
	personRows := make([][]interface{}, 0, len(persons))
	for _, p := range persons {
		personRows = append(personRows, []interface{}{p.Name, p.Contact, p.Unit})
	}
	if err := writeSheetRows(f, "人员", personRows); err != nil {
		return nil, err
	}

	// public_ips
	publicIPRows := make([][]interface{}, 0, len(publicIPs))
	for _, r := range publicIPs {
		publicIPRows = append(publicIPRows, []interface{}{r.IP, r.ISP, r.ExitLocation, r.Remark})
	}
	if err := writeSheetRows(f, "公网IP", publicIPRows); err != nil {
		return nil, err
	}

	// cloud_resources
	cloudRows := make([][]interface{}, 0, len(cloudResources))
	for _, r := range cloudResources {
		cloudRows = append(cloudRows, []interface{}{
			r.Region, r.PhysicalCPU, r.Vcpu, r.Memory, r.Storage, r.BareMetal, r.GpuCardCount, r.ObjectStorage,
		})
	}
	if err := writeSheetRows(f, "云资源", cloudRows); err != nil {
		return nil, err
	}

	// ip_subnets
	subnetRows := make([][]interface{}, 0, len(subnets))
	for _, r := range subnets {
		subnetRows = append(subnetRows, []interface{}{r.CIDR})
	}
	if err := writeSheetRows(f, "网段", subnetRows); err != nil {
		return nil, err
	}

	// hosts（不含 id/ip_mapped/created_at；人员用姓名+联系方式关联）
	hostRows := make([][]interface{}, 0, len(hosts))
	for _, h := range hosts {
		isDB := "否"
		if h.IsDBServer {
			isDB = "是"
		}
		isExternal := "否"
		if h.IsExternal {
			isExternal = "是"
		}
		personName, personContact := "", ""
		if h.Person != nil {
			personName = h.Person.Name
			personContact = h.Person.Contact
		}
		hostRows = append(hostRows, []interface{}{
			h.Region, h.InstanceID, h.Name, h.PrivateIP, h.AssetType, h.OS,
			h.CPU, h.CPUArch, h.Memory, h.ObjectStorage, h.SystemDisk, h.DataDisk,
			h.EnvType, isDB, isExternal, h.Status, h.OpenPorts, h.Tags,
			personName, personContact,
		})
	}
	if err := writeSheetRows(f, "主机", hostRows); err != nil {
		return nil, err
	}

	// host_applications（宿主用内网IP关联）
	appRows := make([][]interface{}, 0, len(hosts))
	for _, h := range hosts {
		a := h.Application
		if a == nil {
			continue
		}
		appRows = append(appRows, []interface{}{
			h.PrivateIP, a.ApplyUnit, a.Applicant, a.ApplicantContact, a.Project,
			a.ApplyReason, a.ApplyConfig, a.ApplyTime, a.ObjectStorageSize, a.Remark,
		})
	}
	if err := writeSheetRows(f, "主机申请信息", appRows); err != nil {
		return nil, err
	}

	// zero_trusts（T-073：按「账户名 + 内网IP」一行导出，同一主机的端口按协议分别并集到
	// 「tcp端口」「udp端口」两列；同账户名不同内网IP 各占一行，其余字段取首条非空。
	// T-077：分组键再带上接入地区——同账户名在不同地区的资源不同，不得合并成同一行）
	ztRows := make([][]interface{}, 0, len(zeroTrusts))
	rowIdx := make(map[string]int, len(zeroTrusts))
	for _, z := range zeroTrusts {
		groups := groupZeroTrustTargets(ztPairs[z.ID], hostIPByID)
		locKey := exitLocationKey(z.ExitLocation)
		if len(groups) == 0 {
			key := z.AccountName + "|" + locKey + "|"
			if idx, ok := rowIdx[key]; ok {
				ztRows[idx] = fillZeroTrustRow(ztRows[idx], zeroTrustRow(z, "", "", ""))
				continue
			}
			rowIdx[key] = len(ztRows)
			ztRows = append(ztRows, zeroTrustRow(z, "", "", ""))
			continue
		}
		for _, g := range groups {
			key := z.AccountName + "|" + locKey + "|" + g.IP
			row := zeroTrustRow(z, g.IP, g.TCP, g.UDP)
			if idx, ok := rowIdx[key]; ok {
				ztRows[idx] = fillZeroTrustRow(ztRows[idx], row)
				continue
			}
			rowIdx[key] = len(ztRows)
			ztRows = append(ztRows, row)
		}
	}
	if err := writeSheetRows(f, "零信任", ztRows); err != nil {
		return nil, err
	}

	// port_mappings（宿主用内网IP关联；isp/exit_location 为派生列不导出）
	mappingRows := make([][]interface{}, 0, len(mappings))
	for _, m := range mappings {
		hostIP := ""
		if m.Host != nil {
			hostIP = m.Host.PrivateIP
		}
		mappingRows = append(mappingRows, []interface{}{
			m.PublicIP, hostIP, m.ExternalPorts, m.InternalPorts, m.ExternalProtocols, m.InternalProtocols, m.Domain, m.Remark,
		})
	}
	if err := writeSheetRows(f, "端口映射", mappingRows); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成导出文件失败: %w", err)
	}
	return buf.Bytes(), nil
}

// writeSheetRows 将数据行写入指定 sheet（表头取自 dataExchangeSheets 定义）
func writeSheetRows(f *excelize.File, sheet string, rows [][]interface{}) error {
	spec := sheetSpecFor(sheet)
	for c, h := range spec.Headers {
		cell, err := excelize.CoordinatesToCellName(c+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
	}
	for r, row := range rows {
		for c, v := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func sheetSpecFor(name string) sheetSpec {
	for _, spec := range dataExchangeSheets {
		if spec.Name == name {
			return spec
		}
	}
	return sheetSpec{}
}

// ExportFilename 导出文件名 mcloud_backup_YYYY-MM-DD.xlsx（日期统一 YYYY-MM-DD）
func ExportFilename() string {
	return "mcloud_backup_" + time.Now().Format("2006-01-02") + ".xlsx"
}

// zeroTrustGroup 单主机（内网IP）的 tcp/udp 端口并集（导出用）
type zeroTrustGroup struct {
	IP  string
	TCP string
	UDP string
}

// groupZeroTrustTargets 接入目标按内网IP 聚合：各协议端口取并集后规范化为文本；
// 同组内 tcp/udp 分别合并（`*` 与其他区间混用时归一）；主机不存在（无内网IP）的目标跳过
func groupZeroTrustTargets(pairs []ZeroTrustTarget, hostIPByID map[uint]string) []zeroTrustGroup {
	order := make([]string, 0, len(pairs))
	tcpRaw := make(map[string][]string, len(pairs))
	udpRaw := make(map[string][]string, len(pairs))
	for _, p := range pairs {
		ip, ok := hostIPByID[p.HostID]
		if !ok {
			continue
		}
		if _, seen := tcpRaw[ip]; !seen {
			if _, seenUDP := udpRaw[ip]; !seenUDP {
				order = append(order, ip)
			}
		}
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		if proto == "udp" {
			udpRaw[ip] = append(udpRaw[ip], p.Ports)
			continue
		}
		tcpRaw[ip] = append(tcpRaw[ip], p.Ports)
	}

	groups := make([]zeroTrustGroup, 0, len(order))
	for _, ip := range order {
		g := zeroTrustGroup{IP: ip}
		if len(tcpRaw[ip]) > 0 {
			if v, err := unionPortText("tcp端口", tcpRaw[ip]...); err == nil {
				g.TCP = v
			}
		}
		if len(udpRaw[ip]) > 0 {
			if v, err := unionPortText("udp端口", udpRaw[ip]...); err == nil {
				g.UDP = v
			}
		}
		groups = append(groups, g)
	}
	return groups
}

// 「零信任」sheet 端口列下标（列序见 dataExchangeSheets，T-073）
const (
	zeroTrustIPCol  = 6
	zeroTrustTCPCol = 7
	zeroTrustUDPCol = 8
)

// zeroTrustRow 组装「零信任」sheet 单行（列序见 dataExchangeSheets）
func zeroTrustRow(z models.ZeroTrust, ip, tcpPorts, udpPorts string) []interface{} {
	return []interface{}{
		z.ApplyUnit, z.AccountName, z.AccountNameZh, z.Applicant, z.Contact, z.ExitLocation,
		ip, tcpPorts, udpPorts, z.SystemName, z.ApplyTime, z.Remark,
	}
}

// fillZeroTrustRow 同账户名 + 内网IP 多行合并：端口取并集，其余字段空位由新行补录
func fillZeroTrustRow(oldRow, newRow []interface{}) []interface{} {
	merged := make([]interface{}, len(oldRow))
	copy(merged, oldRow)
	for i := range merged {
		ov, _ := merged[i].(string)
		nv, _ := newRow[i].(string)
		if nv == "" {
			continue
		}
		if i == zeroTrustTCPCol || i == zeroTrustUDPCol {
			if v, err := unionPortText("端口", ov, nv); err == nil {
				merged[i] = v
			}
			continue
		}
		if ov == "" {
			merged[i] = nv
		}
	}
	return merged
}
