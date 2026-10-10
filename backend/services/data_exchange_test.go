package services

import (
	"testing"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
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
		ApplyUnit: "回灌单位", AccountName: "rt-account", Contact: "张三",
		PublicIP:   "203.0.113.10",
		SystemName: "rt-sys", Remark: "回灌备注",
		ApplyTime: &utils.FlexibleTime{Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		Targets:   []ZeroTrustTarget{{HostID: hostA, Port: 22}, {HostID: hostB, Port: 3389}},
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
	ztStat := report.stat("zero_trusts")
	if ztStat == nil || ztStat.Created != 1 || ztStat.Updated != 0 || ztStat.Skipped != 0 {
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
	if zt.Contact != "张三" || zt.PublicIP != "203.0.113.10" || zt.SystemName != "rt-sys" || zt.Remark != "回灌备注" {
		t.Fatalf("零信任字段回灌不符: %+v", zt)
	}
	if !zt.ApplyTime.Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("apply_time 回灌不符: %v", zt.ApplyTime)
	}
	_ = srcID

	rows := ztTargetRows(t, zt.ID)
	if len(rows) != 2 {
		t.Fatalf("子表期望 2 行，实际 %+v", rows)
	}
	wantPairs := [][2]interface{}{
		{hostIDs[0], 22},
		{hostIDs[1], 3389},
	}
	for i, want := range wantPairs {
		if rows[i].ZeroTrustID != zt.ID || rows[i].HostID != want[0].(uint) || rows[i].Port != want[1].(int) {
			t.Fatalf("子表行%d不符: %+v", i, rows[i])
		}
	}

	// 派生字段按目标库自身主机 ID 组装
	got, err := s.GetByID(zt.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets := joinTargetPairs([]ZeroTrustTarget{
		{HostID: hostIDs[0], Port: 22}, {HostID: hostIDs[1], Port: 3389},
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
	ztStat2 := report2.stat("zero_trusts")
	if ztStat2 == nil || ztStat2.Created != 0 || ztStat2.Updated != 0 || ztStat2.Skipped != 1 {
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
