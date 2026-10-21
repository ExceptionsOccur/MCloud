package services

import (
	"testing"

	"mcloud/database"
	"mcloud/models"
)

// TestZeroTrustXLSXRoundtrip 空库回灌：导出 xlsx → 导入全新库，
// 零信任台账与接入目标子表（zero_trust_targets）完整回灌，二次导入幂等（全部 skip）
func TestZeroTrustXLSXRoundtrip(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hs := NewHostService()
	hostA := makeHost(t, hs, "rt-alpha", "10.67.1.1")
	hostB := makeHost(t, hs, "rt-beta", "10.67.1.2")
	if err := database.DB.Create(&models.PublicIP{IP: "203.0.113.10", ISP: "电信", ExitLocation: "上海"}).Error; err != nil {
		t.Fatalf("种子公网IP失败: %v", err)
	}
	srcID, seedErr := s.Create(ZeroTrustRequest{
		ApplyUnit: "回灌单位", AccountName: "rt-account", AccountNameZh: "回灌账号", Applicant: "回灌人",
		Contact:      "张三",
		ExitLocation: "上海",
		SystemName:   "rt-sys", Remark: "回灌备注",
		ApplyTime: "2026-01-02",
		// T-073：hostA 为 tcp 多端口表达式（导出「tcp端口」列 22,1000-10005），
		// hostB 为 udp（导出「udp端口」列 3389）
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22,1000-10005"}, {HostID: hostB, Ports: "3389", Protocol: "udp"}},
	}, testOp, NewRequestID())
	if seedErr != nil {
		t.Fatalf("种子零信任失败: %v", seedErr)
	}

	data, expErr := NewDataExchangeService().Export()
	if expErr != nil {
		t.Fatalf("导出失败: %v", expErr)
	}

	// 空目标库
	openTestDB(t, t.Name()+"_target")
	importer := NewDataExchangeService()
	report, impErr := importer.Import(data, testOp, NewRequestID())
	if impErr != nil {
		t.Fatalf("导入失败: %v", impErr)
	}
	if !report.Committed {
		t.Fatalf("导入被回滚: %+v", report.Errors)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("导入行错误: %+v", report.Errors)
	}
	// 导出按「每个接入目标一行」→ 2 行；导入按账户名合并：首行 create、次行合并 update
	ztStat := report.stat("零信任")
	if ztStat == nil || ztStat.Rows != 2 || ztStat.Created != 1 || ztStat.Updated != 1 || ztStat.Skipped != 0 {
		t.Fatalf("zero_trusts 导入统计不符: %+v", ztStat)
	}

	// 目标库主机获得新 ID，按自然键（内网IP）取回
	ipA, ipB := "10.67.1.1", "10.67.1.2"
	var hostIDs [2]uint
	for i, ip := range []string{ipA, ipB} {
		var h models.Host
		if err := database.DB.Where("private_ip = ?", ip).First(&h).Error; err != nil {
			t.Fatalf("目标库查主机 %s 失败: %v", ip, err)
		}
		hostIDs[i] = h.ID
	}

	var zt models.ZeroTrust
	if err := database.DB.Where("apply_unit = ? AND account_name = ?", "回灌单位", "rt-account").First(&zt).Error; err != nil {
		t.Fatalf("目标库查零信任失败: %v", err)
	}
	if zt.Applicant != "回灌人" || zt.AccountNameZh != "回灌账号" || zt.Contact != "张三" || zt.ExitLocation != "上海" || zt.SystemName != "rt-sys" || zt.Remark != "回灌备注" {
		t.Fatalf("零信任字段回灌不符: %+v", zt)
	}
	if zt.ApplyTime != "2026-01-02" {
		t.Fatalf("apply_time 回灌不符: %v", zt.ApplyTime)
	}
	_ = srcID

	rows := ztTargetRows(t, zt.ID)
	if len(rows) != 2 {
		t.Fatalf("子表期望 2 行，实际 %+v", rows)
	}
	wantPorts := []string{"22,1000-10005", "3389"}
	wantProtos := []string{"tcp", "udp"}
	for i, want := range wantPorts {
		if rows[i].ZeroTrustID != zt.ID || rows[i].HostID != hostIDs[i] || rows[i].Ports != want || rows[i].Protocol != wantProtos[i] {
			t.Fatalf("子表行%d不符: %+v", i, rows[i])
		}
	}

	// 派生字段按目标库自身主机 ID 组装（T-056 对间以 ; 分隔；T-065 含协议）
	got, err := s.GetByID(zt.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets := joinTargetPairs([]ZeroTrustTarget{
		{HostID: hostIDs[0], Ports: "22,1000-10005"}, {HostID: hostIDs[1], Ports: "3389", Protocol: "udp"},
	})
	if got.Targets != wantTargets {
		t.Fatalf("派生 targets 期望 %q，实际 %q", wantTargets, got.Targets)
	}

	// 二次导入：自然键命中且内容一致 → 全部 skip，不重复建行
	report2, err := importer.Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("二次导入失败: %v", err)
	}
	if !report2.Committed || len(report2.Errors) != 0 {
		t.Fatalf("二次导入异常: committed=%v errors=%+v", report2.Committed, report2.Errors)
	}
	ztStat2 := report2.stat("零信任")
	if ztStat2 == nil || ztStat2.Rows != 2 || ztStat2.Created != 0 || ztStat2.Updated != 0 || ztStat2.Skipped != 2 {
		t.Fatalf("二次导入统计不符: %+v", ztStat2)
	}
	var n int64
	if err := database.DB.Model(&models.ZeroTrustTarget{}).Count(&n).Error; err != nil {
		t.Fatalf("子表计数失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("二次导入后子表应仍为 2 行，实际 %d", n)
	}
}

// TestZeroTrustXLSXMergeByAccount xlsx 导入按账户名合并（T-058）：
// 自然键=账户名（非三元组），目标库已有同账户名记录 → 行并入：
// 字段空补录/非空保留、接入目标并集，不新增记录
func TestZeroTrustXLSXMergeByAccount(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hs := NewHostService()
	srcHost := makeHost(t, hs, "mgx-src", "10.68.1.1")
	if _, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "乙单位", AccountName: "mgx2", Contact: "王五", SystemName: "系统乙",
		Targets: []ZeroTrustTarget{{HostID: srcHost, Ports: "443"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("种子零信任失败: %v", err)
	}
	data, expErr := NewDataExchangeService().Export()
	if expErr != nil {
		t.Fatalf("导出失败: %v", expErr)
	}

	// 目标库：同账户名、不同单位/系统、联系方式为空；文件引用的主机与已有目标主机均在
	openTestDB(t, t.Name()+"_target")
	makeHost(t, hs, "mgx-file", "10.68.1.1")
	dstHost := makeHost(t, hs, "mgx-dst", "10.68.1.2")
	dstID, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "甲单位", AccountName: "mgx2", SystemName: "系统甲",
		Targets: []ZeroTrustTarget{{HostID: dstHost, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("目标库种子失败: %v", err)
	}

	report, impErr := NewDataExchangeService().Import(data, testOp, NewRequestID())
	if impErr != nil {
		t.Fatalf("导入失败: %v", impErr)
	}
	if !report.Committed || len(report.Errors) != 0 {
		t.Fatalf("导入被回滚或有行错误: committed=%v errors=%+v", report.Committed, report.Errors)
	}
	ztStat := report.stat("零信任")
	if ztStat == nil || ztStat.Created != 0 || ztStat.Updated != 1 || ztStat.Skipped != 0 {
		t.Fatalf("zero_trusts 合并统计不符: %+v", ztStat)
	}

	var n int64
	if err = database.DB.Model(&models.ZeroTrust{}).Where("account_name = ?", "mgx2").Count(&n).Error; err != nil {
		t.Fatalf("计数失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("同账户名应仅 1 条，实际 %d", n)
	}
	var rec models.ZeroTrust
	if err = database.DB.First(&rec, dstID).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ApplyUnit != "甲单位" || rec.SystemName != "系统甲" || rec.Contact != "王五" {
		t.Fatalf("合并字段不符（非空保留+空补录）: %+v", rec)
	}
	rows := ztTargetRows(t, dstID)
	if len(rows) != 2 {
		t.Fatalf("targets 应并集为 2 行，实际 %+v", rows)
	}
	fileHost, err := resolveHostIP(database.DB, "10.68.1.1")
	if err != nil {
		t.Fatalf("解析文件主机失败: %v", err)
	}
	if rows[0].HostID != dstHost || rows[0].Ports != "22" || rows[1].HostID != fileHost || rows[1].Ports != "443" {
		t.Fatalf("targets 并集顺序不符: %+v", rows)
	}
}

// TestPortMappingXLSXProtocolRoundtrip 端口映射协议列随 xlsx 往返：
// 导出含外网/内网协议 → 空库回灌保留 → 二次导入按含协议自然键 skip（T-055）
func TestPortMappingXLSXProtocolRoundtrip(t *testing.T) {
	newTestDB(t)
	hostID := makeHost(t, NewHostService(), "pm-rt", "10.67.2.1")
	if err := database.DB.Create(&models.PublicIP{IP: "203.0.113.11", ISP: "联通", ExitLocation: "北京"}).Error; err != nil {
		t.Fatalf("种子公网IP失败: %v", err)
	}
	if _, err := NewPortMappingService().Create(PortMappingRequest{
		PublicIP: "203.0.113.11", HostID: hostID,
		ExternalPorts: "53,1000-10005", InternalPorts: "5353,1000-10005",
		ExternalProtocols: "udp,tcp", InternalProtocols: "udp,tcp",
		Remark: "协议回灌",
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("种子映射失败: %v", err)
	}

	data, expErr := NewDataExchangeService().Export()
	if expErr != nil {
		t.Fatalf("导出失败: %v", expErr)
	}

	openTestDB(t, t.Name()+"_target")
	importer := NewDataExchangeService()
	report, impErr := importer.Import(data, testOp, NewRequestID())
	if impErr != nil {
		t.Fatalf("导入失败: %v", impErr)
	}
	if !report.Committed || len(report.Errors) != 0 {
		t.Fatalf("导入被回滚或有行错误: %+v", report.Errors)
	}
	pmStat := report.stat("端口映射")
	if pmStat == nil || pmStat.Created != 1 || pmStat.Updated != 0 || pmStat.Skipped != 0 {
		t.Fatalf("port_mappings 导入统计不符: %+v", pmStat)
	}

	var m models.PortMapping
	if err := database.DB.First(&m).Error; err != nil {
		t.Fatalf("目标库查映射失败: %v", err)
	}
	if m.ExternalPorts != "53,1000-10005" || m.InternalPorts != "5353,1000-10005" {
		t.Fatalf("端口回灌不符: %q / %q", m.ExternalPorts, m.InternalPorts)
	}
	if m.ExternalProtocols != "udp,tcp" || m.InternalProtocols != "udp,tcp" {
		t.Fatalf("协议回灌不符: %q / %q", m.ExternalProtocols, m.InternalProtocols)
	}

	// 二次导入：自然键含外网协议，内容一致 → skip
	report2, err2 := importer.Import(data, testOp, NewRequestID())
	if err2 != nil {
		t.Fatalf("二次导入失败: %v", err2)
	}
	if !report2.Committed || len(report2.Errors) != 0 {
		t.Fatalf("二次导入异常: committed=%v errors=%+v", report2.Committed, report2.Errors)
	}
	pmStat2 := report2.stat("端口映射")
	if pmStat2 == nil || pmStat2.Created != 0 || pmStat2.Updated != 0 || pmStat2.Skipped != 1 {
		t.Fatalf("二次导入统计不符: %+v", pmStat2)
	}
}
