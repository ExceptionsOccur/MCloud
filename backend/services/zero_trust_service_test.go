package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

// makeHost 按指定名称+内网IP建主机（名称与 IP 解耦，供按主机搜索用例区分命中来源）
func makeHost(t *testing.T, s *HostService, name, ip string) uint {
	t.Helper()
	id, err := s.Create(CreateHostRequest{
		Region:    "华东",
		Name:      name,
		PrivateIP: ip,
		AssetType: "虚拟机",
		EnvType:   "生产",
		Status:    "运行中",
		ApplyTime: "2026-01-02",
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("种子主机失败: %v", err)
	}
	return id
}

// ztTargetRows 读取指定零信任记录的子表行（按行 id 升序）
func ztTargetRows(t *testing.T, zeroTrustID uint) []models.ZeroTrustTarget {
	t.Helper()
	var rows []models.ZeroTrustTarget
	if err := database.DB.Where("zero_trust_id = ?", zeroTrustID).Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatalf("查询子表失败: %v", err)
	}
	return rows
}

// TestZeroTrustTargetSubtableCRUD 配对拆子表：写入/替换/反查/删除全链路 + zero_trusts 无 targets 列
func TestZeroTrustTargetSubtableCRUD(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hs := NewHostService()
	hostA := makeHost(t, hs, "zt-a", "10.66.0.1")
	hostB := makeHost(t, hs, "zt-b", "10.66.0.2")

	// zero_trusts 表结构不残留 targets 列（列已随迁移删除，模型为 gorm:"-" 派生字段）
	var cols []struct{ Name string }
	if err := database.DB.Raw("SELECT name FROM pragma_table_info('zero_trusts')").Scan(&cols).Error; err != nil {
		t.Fatalf("读取表结构失败: %v", err)
	}
	for _, c := range cols {
		if c.Name == "targets" {
			t.Fatal("zero_trusts 表不应存在 targets 列")
		}
	}

	applyTime := &utils.FlexibleTime{Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "测试单位", AccountName: "acct", Contact: "张三",
		SystemName: "sys", ApplyTime: applyTime,
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}, {HostID: hostB, Ports: "443"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	rows := ztTargetRows(t, id)
	if len(rows) != 2 {
		t.Fatalf("子表期望 2 行，实际 %d", len(rows))
	}
	if rows[0].HostID != hostA || rows[0].Ports != "22" || rows[0].ZeroTrustID != id {
		t.Fatalf("子表行0不符: %+v", rows[0])
	}
	if rows[1].HostID != hostB || rows[1].Ports != "443" || rows[1].ZeroTrustID != id {
		t.Fatalf("子表行1不符: %+v", rows[1])
	}

	got, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets := fmt.Sprintf("%d:22;%d:443", hostA, hostB)
	if got.Targets != wantTargets {
		t.Fatalf("派生 targets 期望 %q，实际 %q", wantTargets, got.Targets)
	}

	// Update：整组替换为单条新配对
	err = s.Update(id, ZeroTrustRequest{
		ApplyUnit: "测试单位", AccountName: "acct", Contact: "李四",
		SystemName: "sys", ApplyTime: applyTime,
		Targets: []ZeroTrustTarget{{HostID: hostB, Ports: "8080"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	rows = ztTargetRows(t, id)
	if len(rows) != 1 || rows[0].HostID != hostB || rows[0].Ports != "8080" {
		t.Fatalf("更新后子表不符: %+v", rows)
	}
	got, err = s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets = fmt.Sprintf("%d:8080", hostB)
	if got.Targets != wantTargets {
		t.Fatalf("更新后派生 targets 期望 %q，实际 %q", wantTargets, got.Targets)
	}

	// 按主机反查（派生字段回填）+ 引用计数按记录去重
	zs, err := ZeroTrustsByHost(hostB)
	if err != nil {
		t.Fatalf("ZeroTrustsByHost: %v", err)
	}
	if len(zs) != 1 || zs[0].ID != id {
		t.Fatalf("按主机反查不符: %+v", zs)
	}
	if zs[0].Targets != wantTargets {
		t.Fatalf("反查派生 targets 期望 %q，实际 %q", wantTargets, zs[0].Targets)
	}
	if n, err := HostReferencedByZeroTrust(hostB); err != nil || n != 1 {
		t.Fatalf("引用计数期望 1/nil，实际 %d/%v", n, err)
	}
	if n, err := HostReferencedByZeroTrust(hostA); err != nil || n != 0 {
		t.Fatalf("已替换主机引用计数期望 0/nil，实际 %d/%v", n, err)
	}

	// Delete：子表行随记录删除
	if err := s.Delete(id, testOp, NewRequestID()); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if rows := ztTargetRows(t, id); len(rows) != 0 {
		t.Fatalf("删除后子表应为空，实际 %+v", rows)
	}
}

// TestZeroTrustListSearchByHost 按主机名/内网IP 搜索零信任台账（子表 JOIN hosts）
func TestZeroTrustListSearchByHost(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hs := NewHostService()
	linked := makeHost(t, hs, "unique-alpha-srv", "10.66.9.1")
	makeHost(t, hs, "unique-beta-srv", "10.66.9.2")

	if _, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "搜索单位", AccountName: "search-acct",
		Targets: []ZeroTrustTarget{{HostID: linked, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	for _, tc := range []struct {
		keyword string
		want    int
	}{
		{"unique-alpha-srv", 1}, // 引用主机名
		{"10.66.9.1", 1},        // 引用主机内网IP
		{"search-acct", 1},      // 申请方字段
		{"unique-beta-srv", 0},  // 未被引用的主机名
		{"10.66.9.2", 0},        // 未被引用的主机内网IP
		{"no-such-keyword", 0},
	} {
		items, err := s.List(tc.keyword)
		if err != nil {
			t.Fatalf("List(%q): %v", tc.keyword, err)
		}
		if len(items) != tc.want {
			t.Fatalf("List(%q) 期望 %d 条，实际 %d 条", tc.keyword, tc.want, len(items))
		}
		if tc.want == 1 && len(items[0].Targets) != 1 {
			t.Fatalf("List(%q) 目标视图期望 1 条，实际 %+v", tc.keyword, items[0].Targets)
		}
	}
}

// TestZeroTrustMultiPortExpression 单主机多端口表达式：范围/列表混合、旧 port 数字入参兼容、
// 非法表达式报错、派生 targets 以 ; 分隔对间、List 视图 ports 字段（T-056）
func TestZeroTrustMultiPortExpression(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "zt-mp", "10.66.8.1")

	legacyPort := 22
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "多端口单位", AccountName: "mp-acct",
		Targets: []ZeroTrustTarget{
			{HostID: hostA, Ports: "1000-10005,80"},
			{HostID: hostA, Port: &legacyPort},
		},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	rows := ztTargetRows(t, id)
	if len(rows) != 2 {
		t.Fatalf("子表期望 2 行，实际 %+v", rows)
	}
	if rows[0].Ports != "1000-10005,80" {
		t.Fatalf("表达式落库不符: %q", rows[0].Ports)
	}
	if rows[1].Ports != "22" {
		t.Fatalf("旧 port 入参应归一为单口表达式: %q", rows[1].Ports)
	}

	got, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets := fmt.Sprintf("%d:1000-10005,80;%d:22", hostA, hostA)
	if got.Targets != wantTargets {
		t.Fatalf("派生 targets 期望 %q，实际 %q", wantTargets, got.Targets)
	}

	// 非法范围表达式 → 校验错误（40001 路径）
	if _, err = s.Create(ZeroTrustRequest{
		ApplyUnit: "多端口单位", AccountName: "mp-bad",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "10005-1000"}},
	}, testOp, NewRequestID()); err == nil || !strings.Contains(err.Error(), "不合法端口范围") {
		t.Fatalf("期望范围校验错误，实际 %v", err)
	}

	// List 视图携带 ports 表达式
	items, err := s.List("")
	if err != nil || len(items) != 1 {
		t.Fatalf("List: %v / %d 条", err, len(items))
	}
	if len(items[0].Targets) != 2 || items[0].Targets[0].Ports != "1000-10005,80" {
		t.Fatalf("List 目标视图不符: %+v", items[0].Targets)
	}
}

// TestZeroTrustBatchPortGroups 批量添加端口组：引号包裹多端口列表、旧平铺兼容、
// IP/端口组数量不一致行错误、非法表达式行错误（T-056）
func TestZeroTrustBatchPortGroups(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	makeHost(t, hs, "bg-a", "10.66.7.1")
	makeHost(t, hs, "bg-b", "10.66.7.2")
	s := NewZeroTrustService()

	text := strings.Join([]string{
		`批量单位,acct-a,,"10.66.7.1,10.66.7.2","""1000-10005,22"",8080",系统A,,,`,
		`批量单位,acct-b,,10.66.7.1,443,系统B,,,`,
		`批量单位,acct-c,,"10.66.7.1,10.66.7.2",55,系统C,,,`,
		`批量单位,acct-d,,10.66.7.1,10005-1000,系统D,,,`,
	}, "\n")
	resp, err := s.BatchCreateText(text, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量添加失败: %v", err)
	}
	if resp.Success != 2 || resp.Errors != 2 {
		t.Fatalf("统计期望 成功2/错误2，实际 成功%d/错误%d：%+v", resp.Success, resp.Errors, resp.LineErrors)
	}

	byAccount := make(map[string][]string)
	var records []models.ZeroTrust
	if err := database.DB.Order("id ASC").Find(&records).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	for _, r := range records {
		pairs, err := loadZeroTrustTargetPairs(database.DB, []uint{r.ID})
		if err != nil {
			t.Fatalf("读子表失败: %v", err)
		}
		for _, p := range pairs[r.ID] {
			byAccount[r.AccountName] = append(byAccount[r.AccountName], p.Ports)
		}
	}
	if got := byAccount["acct-a"]; len(got) != 2 || got[0] != "1000-10005,22" || got[1] != "8080" {
		t.Fatalf("acct-a 端口组不符: %v", got)
	}
	if got := byAccount["acct-b"]; len(got) != 1 || got[0] != "443" {
		t.Fatalf("acct-b 旧平铺兼容不符: %v", got)
	}
	if _, ok := byAccount["acct-c"]; ok {
		t.Fatal("acct-c 数量不一致不应入库")
	}
	if _, ok := byAccount["acct-d"]; ok {
		t.Fatal("acct-d 非法表达式不应入库")
	}
}
