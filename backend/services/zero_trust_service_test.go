package services

import (
	"fmt"
	"strings"
	"testing"

	"mcloud/database"
	"mcloud/models"
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

	applyTime := "2026-01-02"
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
	wantTargets := fmt.Sprintf("%d:22:tcp;%d:443:tcp", hostA, hostB)
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
	wantTargets = fmt.Sprintf("%d:8080:tcp", hostB)
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
		ApplyUnit: "搜索单位", AccountName: "search-acct", Applicant: "unique-applicant-x",
		Targets: []ZeroTrustTarget{{HostID: linked, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	for _, tc := range []struct {
		keyword string
		want    int
	}{
		{"unique-alpha-srv", 1},   // 引用主机名
		{"10.66.9.1", 1},          // 引用主机内网IP
		{"search-acct", 1},        // 账户名字段
		{"unique-applicant-x", 1}, // 申请人字段（T-060）
		{"unique-beta-srv", 0},    // 未被引用的主机名
		{"10.66.9.2", 0},          // 未被引用的主机内网IP
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
	wantTargets := fmt.Sprintf("%d:1000-10005,80:tcp;%d:22:tcp", hostA, hostA)
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

// TestZeroTrustTargetProtocol T-065 接入目标协议（每行单协议，缺省 tcp）：
// Create/Update 落库回显、默认 tcp、非法协议报错、派生串含协议、批量 `/协议` 后缀、单主机多端口协议不一致报错
func TestZeroTrustTargetProtocol(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "zp-a", "10.66.7.9")
	hostB := makeHost(t, NewHostService(), "zp-b", "10.66.7.10")

	// 表结构含 protocol 列
	var cols []struct{ Name string }
	if err := database.DB.Raw("SELECT name FROM pragma_table_info('zero_trust_targets')").Scan(&cols).Error; err != nil {
		t.Fatalf("读取表结构失败: %v", err)
	}
	hasProto := false
	for _, c := range cols {
		if c.Name == "protocol" {
			hasProto = true
		}
	}
	if !hasProto {
		t.Fatal("zero_trust_targets 应含 protocol 列")
	}

	// Create：显式 udp + 缺省 tcp
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "协议单位", AccountName: "proto-acct",
		Targets: []ZeroTrustTarget{
			{HostID: hostA, Ports: "53", Protocol: "udp"},
			{HostID: hostB, Ports: "443"},
		},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	rows := ztTargetRows(t, id)
	if len(rows) != 2 || rows[0].Protocol != "udp" || rows[1].Protocol != "tcp" {
		t.Fatalf("协议落库不符（udp/缺省tcp）: %+v", rows)
	}
	got, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	wantTargets := fmt.Sprintf("%d:53:udp;%d:443:tcp", hostA, hostB)
	if got.Targets != wantTargets {
		t.Fatalf("派生 targets 期望 %q，实际 %q", wantTargets, got.Targets)
	}
	items, err := s.List("")
	if err != nil || len(items) != 1 {
		t.Fatalf("List: %v / %d", err, len(items))
	}
	if items[0].Targets[0].Protocol != "udp" || items[0].Targets[1].Protocol != "tcp" {
		t.Fatalf("List 协议视图不符: %+v", items[0].Targets)
	}

	// 非法协议报错
	if _, e := s.Create(ZeroTrustRequest{
		ApplyUnit: "协议单位", AccountName: "proto-bad",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "80", Protocol: "sctp"}},
	}, testOp, NewRequestID()); e == nil {
		t.Fatal("非法协议应报错")
	}

	// 批量：`/协议` 后缀（整个端口组），缺省 tcp
	text := strings.Join([]string{
		`协议单位,bp-1,,,,"10.66.7.9",53/udp,系统1,,,`,
		`协议单位,bp-2,,,,10.66.7.10,80,系统2,,,`,
	}, "\n")
	resp, err := s.BatchCreateText(text, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量失败: %v", err)
	}
	if resp.Success != 2 || resp.Errors != 0 {
		t.Fatalf("批量统计不符: %+v：%+v", resp, resp.LineErrors)
	}
	var bp1 models.ZeroTrust
	if err := database.DB.Where("account_name = ?", "bp-1").First(&bp1).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	r1 := ztTargetRows(t, bp1.ID)
	if len(r1) != 1 || r1[0].Ports != "53" || r1[0].Protocol != "udp" {
		t.Fatalf("批量 udp 组不符: %+v", r1)
	}
}

// TestZeroTrustBatchSemicolonGroups 申请端口列 `;` 分组：多IP各自多端口、
// 单IP+整列逗号免双层引号直接视为多端口、组数与IP数不一致行错误、旧平铺回归（T-058）
func TestZeroTrustBatchSemicolonGroups(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	makeHost(t, hs, "sc-a", "10.66.6.1")
	makeHost(t, hs, "sc-b", "10.66.6.2")
	s := NewZeroTrustService()

	text := strings.Join([]string{
		`分号单位,sc-a,,,,"10.66.6.1,10.66.6.2","22,443;8080,1000-10005",系统A,,,`,
		`分号单位,sc-b,,,,"10.66.6.1","22,443",系统B,,,`,
		`分号单位,sc-c,,,,"10.66.6.1,10.66.6.2","22;443;8080",系统C,,,`,
		`分号单位,sc-d,,,,10.66.6.1,443,系统D,,,`,
	}, "\n")
	resp, err := s.BatchCreateText(text, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量添加失败: %v", err)
	}
	if resp.Success != 3 || resp.Errors != 1 {
		t.Fatalf("统计期望 成功3/错误1，实际 成功%d/错误%d：%+v", resp.Success, resp.Errors, resp.LineErrors)
	}

	got := make(map[string][]string)
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
			got[r.AccountName] = append(got[r.AccountName], p.Ports)
		}
	}
	if v := got["sc-a"]; len(v) != 2 || v[0] != "22,443" || v[1] != "8080,1000-10005" {
		t.Fatalf("sc-a 分号分组不符: %v", v)
	}
	if v := got["sc-b"]; len(v) != 1 || v[0] != "22,443" {
		t.Fatalf("sc-b 单IP整列多端口不符: %v", v)
	}
	if _, ok := got["sc-c"]; ok {
		t.Fatal("sc-c 组数与IP数不一致不应入库")
	}
	if v := got["sc-d"]; len(v) != 1 || v[0] != "443" {
		t.Fatalf("sc-d 旧平铺回归不符: %v", v)
	}
}

// TestZeroTrustCreateMergeByAccount 新增按账户名合并（T-058）：
// 同账户名二次 Create 合并进 id 最小一条，targets 并集去重、空字段补录、非空保留、
// 无变化不重复写库；接入地区仅补录空位、已有值时丢弃新值（T-061）
func TestZeroTrustCreateMergeByAccount(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "mg-a", "10.66.5.1")
	hostB := makeHost(t, NewHostService(), "mg-b", "10.66.5.2")

	id1, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "合并单位", AccountName: "mg-acct",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("首次创建失败: %v", err)
	}

	// 二次新增：申请单位不同（不覆盖）、联系方式空位补录、targets 并集去重
	id2, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "其他单位", AccountName: "mg-acct", Contact: "李四",
		SystemName: "新系统", Remark: "备注补录",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}, {HostID: hostB, Ports: "8080"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("合并创建失败: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("合并应返回已有记录 id：id1=%d id2=%d", id1, id2)
	}

	var n int64
	if err := database.DB.Model(&models.ZeroTrust{}).Where("account_name = ?", "mg-acct").Count(&n).Error; err != nil {
		t.Fatalf("计数失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("同账户名应仅 1 条记录，实际 %d", n)
	}
	var rec models.ZeroTrust
	if err := database.DB.First(&rec, id1).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ApplyUnit != "合并单位" || rec.Contact != "李四" || rec.SystemName != "新系统" || rec.Remark != "备注补录" {
		t.Fatalf("合并字段不符（非空保留+空补录）: %+v", rec)
	}
	rows := ztTargetRows(t, id1)
	if len(rows) != 2 || rows[0].HostID != hostA || rows[0].Ports != "22" || rows[1].HostID != hostB || rows[1].Ports != "8080" {
		t.Fatalf("targets 并集不符: %+v", rows)
	}

	// 完全重复的新增：无变化仍返回同 id，不新增不改行
	if id3, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "合并单位", AccountName: "mg-acct", Contact: "李四",
		SystemName: "新系统", Remark: "备注补录",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}, {HostID: hostB, Ports: "8080"}},
	}, testOp, NewRequestID()); err != nil || id3 != id1 {
		t.Fatalf("重复新增应无变化返回原 id：%d / %v", id3, err)
	}

	// 接入地区：已有记录该字段为空 → 补录
	if _, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "合并单位", AccountName: "mg-acct", ExitLocation: "南宁",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("接入地区补录失败: %v", err)
	}
	if err := database.DB.First(&rec, id1).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ExitLocation != "南宁" {
		t.Fatalf("接入地区补录不符，实际 %q", rec.ExitLocation)
	}
	// 已有接入地区 → 新值并集去重（T-065，非覆盖）
	if _, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "合并单位", AccountName: "mg-acct", ExitLocation: "青年",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("接入地区并集失败: %v", err)
	}
	if err := database.DB.First(&rec, id1).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ExitLocation != "南宁,青年" {
		t.Fatalf("接入地区应并集去重，实际 %q", rec.ExitLocation)
	}
	// 重复地区不再追加
	if _, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "合并单位", AccountName: "mg-acct", ExitLocation: "南宁",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("接入地区重复合并失败: %v", err)
	}
	if err := database.DB.First(&rec, id1).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ExitLocation != "南宁,青年" {
		t.Fatalf("接入地区重复值应去重，实际 %q", rec.ExitLocation)
	}
}

// TestZeroTrustBatchMergeByAccount 批量文本同账户名两行合并为一条（T-058）
func TestZeroTrustBatchMergeByAccount(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	makeHost(t, hs, "bm-a", "10.66.4.1")
	makeHost(t, hs, "bm-b", "10.66.4.2")
	s := NewZeroTrustService()

	text := strings.Join([]string{
		`批量单位,ba-merge,,,,10.66.4.1,22,系统X,,,`,
		`批量单位,ba-merge,,,,10.66.4.2,443,系统Y,,,`,
	}, "\n")
	resp, err := s.BatchCreateText(text, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量添加失败: %v", err)
	}
	if resp.Success != 2 || resp.Errors != 0 {
		t.Fatalf("统计期望 成功2/失败0，实际 %+v：%+v", resp, resp.LineErrors)
	}
	var n int64
	if err := database.DB.Model(&models.ZeroTrust{}).Where("account_name = ?", "ba-merge").Count(&n).Error; err != nil {
		t.Fatalf("计数失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("同账户名两行应合并为 1 条，实际 %d", n)
	}
	var rec models.ZeroTrust
	if err := database.DB.Where("account_name = ?", "ba-merge").First(&rec).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.SystemName != "系统X" {
		t.Fatalf("系统名称非空应保留首行值，实际 %q", rec.SystemName)
	}
	rows := ztTargetRows(t, rec.ID)
	if len(rows) != 2 {
		t.Fatalf("targets 应并集为 2 行，实际 %+v", rows)
	}
}

// TestZeroTrustBatchSingleGroupBroadcast 多IP相同端口（T-059）：无 `;` 单端口组广播到全部 IP、
// 重复分组与 `;` 按位配对回归、`;` 单组+多IP 与组数仍不匹配行报错
func TestZeroTrustBatchSingleGroupBroadcast(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	makeHost(t, hs, "bc-a", "10.66.3.1")
	makeHost(t, hs, "bc-b", "10.66.3.2")
	makeHost(t, hs, "bc-c", "10.66.3.3")
	s := NewZeroTrustService()

	text := strings.Join([]string{
		`广播单位,bc-1,,,,"10.66.3.1,10.66.3.2",22,系统1,,,`,
		`广播单位,bc-2,,,,"10.66.3.1,10.66.3.2","22,443;22,443",系统2,,,`,
		`广播单位,bc-3,,,,"10.66.3.1,10.66.3.2",22;22,系统3,,,`,
		`广播单位,bc-4,,,,"10.66.3.1,10.66.3.2",55;80,系统4,,,`,
		`广播单位,bc-5,,,,"10.66.3.1,10.66.3.2","22;",系统5,,,`,
		`广播单位,bc-6,,,,"10.66.3.1,10.66.3.2,10.66.3.3","22,443",系统6,,,`,
	}, "\n")
	resp, err := s.BatchCreateText(text, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量添加失败: %v", err)
	}
	if resp.Success != 4 || resp.Errors != 2 {
		t.Fatalf("统计期望 成功4/错误2，实际 成功%d/错误%d：%+v", resp.Success, resp.Errors, resp.LineErrors)
	}

	got := make(map[string][]string)
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
			got[r.AccountName] = append(got[r.AccountName], p.Ports)
		}
	}
	if v := got["bc-1"]; len(v) != 2 || v[0] != "22" || v[1] != "22" {
		t.Fatalf("bc-1 单组广播不符: %v", v)
	}
	if v := got["bc-2"]; len(v) != 2 || v[0] != "22,443" || v[1] != "22,443" {
		t.Fatalf("bc-2 重复分组不符: %v", v)
	}
	if v := got["bc-3"]; len(v) != 2 || v[0] != "22" || v[1] != "22" {
		t.Fatalf("bc-3 分号重复不符: %v", v)
	}
	if v := got["bc-4"]; len(v) != 2 || v[0] != "55" || v[1] != "80" {
		t.Fatalf("bc-4 `;` 按位配对不符: %v", v)
	}
	if _, ok := got["bc-5"]; ok {
		t.Fatal("bc-5 `;` 单组+多IP 不应入库")
	}
	if _, ok := got["bc-6"]; ok {
		t.Fatal("bc-6 组数与IP数不一致不应入库")
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
		`批量单位,acct-a,,,,"10.66.7.1,10.66.7.2","""1000-10005,22"",8080",系统A,,,`,
		`批量单位,acct-b,,,,10.66.7.1,443,系统B,,,`,
		`批量单位,acct-c,,,,"10.66.7.1,10.66.7.2","55;80;90",系统C,,,`,
		`批量单位,acct-d,,,,10.66.7.1,10005-1000,系统D,,,`,
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

// TestZeroTrustExitLocationAndSchema T-061：zero_trusts 无 public_ip 列、有 exit_location 列；
// ListExitLocations 返回 public_ips.exit_location 去重非空集；Create 落库回显接入地区
func TestZeroTrustExitLocationAndSchema(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "el-a", "10.66.1.1")

	var cols []struct{ Name string }
	if err := database.DB.Raw("SELECT name FROM pragma_table_info('zero_trusts')").Scan(&cols).Error; err != nil {
		t.Fatalf("读取表结构失败: %v", err)
	}
	hasExit, hasPublic := false, false
	for _, c := range cols {
		if c.Name == "exit_location" {
			hasExit = true
		}
		if c.Name == "public_ip" {
			hasPublic = true
		}
	}
	if !hasExit || hasPublic {
		t.Fatalf("zero_trusts 应含 exit_location 且不含 public_ip（exit=%v public=%v）", hasExit, hasPublic)
	}

	// 地区候选值取资源池去重非空
	for _, r := range []models.PublicIP{
		{IP: "203.0.113.1", ExitLocation: "南宁"},
		{IP: "203.0.113.2", ExitLocation: "南宁"},
		{IP: "203.0.113.3", ExitLocation: "青年"},
		{IP: "203.0.113.4", ExitLocation: ""},
	} {
		if err := database.DB.Create(&r).Error; err != nil {
			t.Fatalf("种子公网IP失败: %v", err)
		}
	}
	locs, err := s.ListExitLocations()
	if err != nil {
		t.Fatalf("ListExitLocations: %v", err)
	}
	if len(locs) != 2 || locs[0] != "南宁" || locs[1] != "青年" {
		t.Fatalf("地区候选去重不符: %v", locs)
	}

	// Create 落库接入地区并回显
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "地区单位", AccountName: "el-acct", ExitLocation: "南宁",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	items, err := s.List("")
	if err != nil || len(items) != 1 {
		t.Fatalf("List: %v / %d 条", err, len(items))
	}
	if items[0].ID != id || items[0].ExitLocation != "南宁" {
		t.Fatalf("接入地区回显不符: %+v", items[0].ZeroTrust)
	}
	// 搜索命中接入地区
	if items, err := s.List("南宁"); err != nil || len(items) != 1 {
		t.Fatalf("按接入地区搜索应命中 1 条: %v / %d", err, len(items))
	}
}

// TestZeroTrustExitLocationMulti T-065 接入地区多值：逗号/顿号分隔规整去重、
// Create/Update 多值落库回显、搜索命中任一地区、合并并集
func TestZeroTrustExitLocationMulti(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "elm-a", "10.66.1.9")

	// 逗号+顿号+重复+空白 → 规整去重保序
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "多地区单位", AccountName: "elm-acct", ExitLocation: " 南宁、青年,南宁 ， 深圳 ",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	var rec models.ZeroTrust
	if err := database.DB.First(&rec, id).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ExitLocation != "南宁,青年,深圳" {
		t.Fatalf("多值规整去重不符: %q", rec.ExitLocation)
	}
	// 搜索命中任一地区
	for _, kw := range []string{"南宁", "青年", "深圳"} {
		if items, err := s.List(kw); err != nil || len(items) != 1 {
			t.Fatalf("按 %s 搜索应命中 1 条: %v / %d", kw, err, len(items))
		}
	}

	// Update 替换为另一组多值
	if err := s.Update(id, ZeroTrustRequest{
		ApplyUnit: "多地区单位", AccountName: "elm-acct", ExitLocation: "北京,上海",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if err := database.DB.First(&rec, id).Error; err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.ExitLocation != "北京,上海" {
		t.Fatalf("更新后多值不符: %q", rec.ExitLocation)
	}
}

// TestZeroTrustAccountNameZh T-063「账号姓名」字段（account_name_zh）：
// Create/Update 独立存储回显、账户名合并空字段补录/非空保留、搜索命中
func TestZeroTrustAccountNameZh(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "anz-a", "10.66.0.9")

	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "账号姓名单位", AccountName: "anz-acct", AccountNameZh: "张三丰",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	_ = err
	var rec models.ZeroTrust
	if e := database.DB.First(&rec, id).Error; e != nil {
		t.Fatalf("查记录失败: %v", e)
	}
	if rec.AccountName != "anz-acct" || rec.AccountNameZh != "张三丰" {
		t.Fatalf("账户名/账号姓名应独立存储: %+v", rec)
	}
	if items, e := s.List("张三丰"); e != nil || len(items) != 1 {
		t.Fatalf("按账号姓名搜索应命中 1 条: %v / %d", e, len(items))
	}

	// 合并：已有账号姓名非空 → 新值丢弃
	if id2, e := s.Create(ZeroTrustRequest{
		ApplyUnit: "账号姓名单位", AccountName: "anz-acct", AccountNameZh: "李四",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "443"}},
	}, testOp, NewRequestID()); e != nil || id2 != id {
		t.Fatalf("合并失败: id=%d err=%v", id2, e)
	}
	if e := database.DB.First(&rec, id).Error; e != nil {
		t.Fatalf("查记录失败: %v", e)
	}
	if rec.AccountNameZh != "张三丰" {
		t.Fatalf("已有账号姓名非空应保留，实际 %q", rec.AccountNameZh)
	}

	// 合并：已有账号姓名空 → 新值补录
	id3, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "账号姓名单位", AccountName: "anz-acct2",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if _, e := s.Create(ZeroTrustRequest{
		ApplyUnit: "账号姓名单位", AccountName: "anz-acct2", AccountNameZh: "王五",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "80"}},
	}, testOp, NewRequestID()); e != nil {
		t.Fatalf("合并补录失败: %v", e)
	}
	var rec2 models.ZeroTrust
	if e := database.DB.First(&rec2, id3).Error; e != nil {
		t.Fatalf("查记录失败: %v", e)
	}
	if rec2.AccountNameZh != "王五" {
		t.Fatalf("空字段应补录账号姓名，实际 %q", rec2.AccountNameZh)
	}

	// Update 独立更新账号姓名
	if e := s.Update(id, ZeroTrustRequest{
		ApplyUnit: "账号姓名单位", AccountName: "anz-acct", AccountNameZh: "赵六",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID()); e != nil {
		t.Fatalf("更新失败: %v", e)
	}
	if e := database.DB.First(&rec, id).Error; e != nil {
		t.Fatalf("查记录失败: %v", e)
	}
	if rec.AccountNameZh != "赵六" {
		t.Fatalf("更新后账号姓名不符: %q", rec.AccountNameZh)
	}
}

// TestZeroTrustApplicant T-060「申请人」字段与账户名区分：
// Create/Update 独立存储并回显、账户名合并时空字段补录/非空保留、批量第 3 列解析
func TestZeroTrustApplicant(t *testing.T) {
	newTestDB(t)
	s := NewZeroTrustService()
	hostA := makeHost(t, NewHostService(), "ap-a", "10.66.2.1")

	// Create：申请人独立于账户名
	id, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "申请人单位", AccountName: "ap-acct", Applicant: "张三",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	var rec models.ZeroTrust
	err = database.DB.First(&rec, id).Error
	if err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.AccountName != "ap-acct" || rec.Applicant != "张三" {
		t.Fatalf("账户名/申请人应独立存储: %+v", rec)
	}

	// 账户名合并：已有申请人非空 → 新值丢弃
	id2, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "申请人单位", AccountName: "ap-acct", Applicant: "李四",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "443"}},
	}, testOp, NewRequestID())
	if err != nil || id2 != id {
		t.Fatalf("合并失败: id=%d err=%v", id2, err)
	}
	err = database.DB.First(&rec, id).Error
	if err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.Applicant != "张三" {
		t.Fatalf("已有申请人非空应保留，实际 %q", rec.Applicant)
	}

	// 账户名合并：已有申请人空 → 新值补录
	id3, err := s.Create(ZeroTrustRequest{
		ApplyUnit: "申请人单位", AccountName: "ap-acct2", Applicant: "",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	_, err = s.Create(ZeroTrustRequest{
		ApplyUnit: "申请人单位", AccountName: "ap-acct2", Applicant: "王五",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "80"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("合并补录失败: %v", err)
	}
	var rec2 models.ZeroTrust
	err = database.DB.First(&rec2, id3).Error
	if err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec2.Applicant != "王五" {
		t.Fatalf("空字段应补录申请人，实际 %q", rec2.Applicant)
	}

	// Update：申请人与账户名独立更新
	err = s.Update(id, ZeroTrustRequest{
		ApplyUnit: "申请人单位", AccountName: "ap-acct", Applicant: "赵六",
		Targets: []ZeroTrustTarget{{HostID: hostA, Ports: "22"}},
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	err = database.DB.First(&rec, id).Error
	if err != nil {
		t.Fatalf("查记录失败: %v", err)
	}
	if rec.Applicant != "赵六" || rec.AccountName != "ap-acct" {
		t.Fatalf("更新后字段不符: %+v", rec)
	}

	// 批量：第 3 列为账号姓名、第 4 列为申请人
	makeHost(t, NewHostService(), "ap-b", "10.66.2.2")
	resp, err := s.BatchCreateText(`批量申请人单位,bp-acct,批量账号姓名,批量张三,138,10.66.2.2,8080,系统,,,`, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("批量添加失败: %v", err)
	}
	if resp.Success != 1 || resp.Errors != 0 {
		t.Fatalf("批量统计不符: %+v：%+v", resp, resp.LineErrors)
	}
	var bp models.ZeroTrust
	err = database.DB.Where("account_name = ?", "bp-acct").First(&bp).Error
	if err != nil {
		t.Fatalf("查批量记录失败: %v", err)
	}
	if bp.AccountNameZh != "批量账号姓名" || bp.Applicant != "批量张三" || bp.Contact != "138" {
		t.Fatalf("批量账号姓名/申请人/联系方式列不符: %+v", bp)
	}
}
