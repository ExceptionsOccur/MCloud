package services

import (
	"fmt"
	"strings"
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
	if err := writeSheetRows(f, "persons", personRows); err != nil {
		return nil, err
	}

	// public_ips
	publicIPRows := make([][]interface{}, 0, len(publicIPs))
	for _, r := range publicIPs {
		publicIPRows = append(publicIPRows, []interface{}{r.IP, r.ISP, r.ExitLocation, r.Remark})
	}
	if err := writeSheetRows(f, "public_ips", publicIPRows); err != nil {
		return nil, err
	}

	// cloud_resources
	cloudRows := make([][]interface{}, 0, len(cloudResources))
	for _, r := range cloudResources {
		cloudRows = append(cloudRows, []interface{}{
			r.Region, r.PhysicalCPU, r.Vcpu, r.Memory, r.Storage, r.BareMetal, r.GpuCardCount, r.ObjectStorage,
		})
	}
	if err := writeSheetRows(f, "cloud_resources", cloudRows); err != nil {
		return nil, err
	}

	// ip_subnets
	subnetRows := make([][]interface{}, 0, len(subnets))
	for _, r := range subnets {
		subnetRows = append(subnetRows, []interface{}{r.CIDR})
	}
	if err := writeSheetRows(f, "ip_subnets", subnetRows); err != nil {
		return nil, err
	}

	// hosts（不含 id/ip_mapped/created_at/disk；人员用姓名+联系方式关联）
	hostRows := make([][]interface{}, 0, len(hosts))
	for _, h := range hosts {
		isDB := "否"
		if h.IsDBServer {
			isDB = "是"
		}
		personName, personContact := "", ""
		if h.Person != nil {
			personName = h.Person.Name
			personContact = h.Person.Contact
		}
		hostRows = append(hostRows, []interface{}{
			h.Region, h.InstanceID, h.Name, h.PrivateIP, h.AssetType, h.OS,
			h.CPU, h.CPUArch, h.Memory, h.SystemDisk, h.DataDisk,
			h.EnvType, isDB, h.Status, h.OpenPorts, h.Tags,
			personName, personContact,
		})
	}
	if err := writeSheetRows(f, "hosts", hostRows); err != nil {
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
	if err := writeSheetRows(f, "host_applications", appRows); err != nil {
		return nil, err
	}

	// zero_trusts（接入目标 host_id:port → 内网IP:端口）
	ztRows := make([][]interface{}, 0, len(zeroTrusts))
	for _, z := range zeroTrusts {
		pairs := parseStoredTargets(z.Targets)
		parts := make([]string, 0, len(pairs))
		for _, p := range pairs {
			ip, ok := hostIPByID[p.HostID]
			if !ok {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s:%d", ip, p.Port))
		}
		ztRows = append(ztRows, []interface{}{
			z.ApplyUnit, z.AccountName, z.Contact, z.PublicIP,
			strings.Join(parts, ","), z.SystemName,
			z.ApplyTime.Format("2006-01-02 15:04:05"), z.Remark,
		})
	}
	if err := writeSheetRows(f, "zero_trusts", ztRows); err != nil {
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
			m.PublicIP, hostIP, m.ExternalPorts, m.InternalPorts, m.Domain, m.Remark,
		})
	}
	if err := writeSheetRows(f, "port_mappings", mappingRows); err != nil {
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

// ExportFilename 导出文件名 mcloud_backup_YYYYMMDD.xlsx
func ExportFilename() string {
	return "mcloud_backup_" + time.Now().Format("20060102") + ".xlsx"
}
