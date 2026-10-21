package services

import (
	"errors"
	"strings"
	"testing"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

func baseHostReq(ip string) CreateHostRequest {
	return CreateHostRequest{
		Region:    "华东",
		Name:      "web-" + ip,
		PrivateIP: ip,
		AssetType: "虚拟机",
		EnvType:   "生产",
		Status:    "运行中",
		ApplyTime: "2026-01-02",
	}
}

var testIPSeq int

func uniqueTestIP() string {
	testIPSeq++
	b := []byte("10.88.0.000")
	n := testIPSeq
	for i := 10; i >= 8 && n > 0; i-- {
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b)
}

func TestHostCreate(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	tests := []struct {
		name       string
		mutate     func(req *CreateHostRequest)
		wantErr    error // 哨兵错误；nil 表示无哨兵要求
		wantErrSub string
	}{
		{
			name:   "成功创建并落申请信息",
			mutate: func(req *CreateHostRequest) { req.Applicant = "张三" },
		},
		{
			name: "内网IP重复拒绝",
			mutate: func(req *CreateHostRequest) {
				mustCreateHost(t, s, "10.0.0.1")
				req.PrivateIP = "10.0.0.1"
			},
			wantErrSub: "内网IP已存在",
		},
		{
			name:       "申请时间非严格格式拒绝",
			mutate:     func(req *CreateHostRequest) { req.ApplyTime = "2026/01/02" },
			wantErrSub: "YYYY-MM-DD",
		},
		{
			name: "人员不存在拒绝",
			mutate: func(req *CreateHostRequest) {
				missing := uint(999)
				req.PersonID = OptionalUint{Present: true, Value: &missing}
			},
			wantErr: ErrPersonNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := baseHostReq(uniqueTestIP())
			tt.mutate(&req)

			id, err := s.Create(req, testOp, NewRequestID())
			if tt.wantErr != nil || tt.wantErrSub != "" {
				if err == nil {
					t.Fatalf("期望错误，实际成功 id=%d", id)
				}
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Fatalf("期望哨兵 %v，实际 %v", tt.wantErr, err)
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("错误信息应含 %q，实际 %q", tt.wantErrSub, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Create 失败: %v", err)
			}
			if id == 0 {
				t.Fatal("Create 返回 id=0")
			}

			var host models.Host
			if err := database.DB.First(&host, id).Error; err != nil {
				t.Fatalf("回查主机失败: %v", err)
			}
			if host.PrivateIP != req.PrivateIP || host.Name != req.Name {
				t.Fatalf("落库字段不符: got ip=%s name=%s", host.PrivateIP, host.Name)
			}
			var app models.HostApplication
			if err := database.DB.Where("host_id = ?", id).First(&app).Error; err != nil {
				t.Fatalf("回查申请信息失败: %v", err)
			}
			if app.Applicant != "张三" {
				t.Fatalf("申请信息未落库: applicant=%q", app.Applicant)
			}
		})
	}
}

func TestHostGetAndUpdate(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	t.Run("GetByID未命中", func(t *testing.T) {
		if _, err := s.GetByID(404); err == nil {
			t.Fatal("期望错误")
		}
	})

	id := mustCreateHost(t, s, "10.1.1.1")

	t.Run("GetByID命中", func(t *testing.T) {
		host, err := s.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID 失败: %v", err)
		}
		if host.Name != "host-10.1.1.1" {
			t.Fatalf("名称不符: %s", host.Name)
		}
	})

	t.Run("Update改名成功", func(t *testing.T) {
		if err := s.Update(id, UpdateHostRequest{Name: "renamed"}, testOp, NewRequestID()); err != nil {
			t.Fatalf("Update 失败: %v", err)
		}
		host, err := s.GetByID(id)
		if err != nil {
			t.Fatalf("回查失败: %v", err)
		}
		if host.Name != "renamed" {
			t.Fatalf("改名未生效: %s", host.Name)
		}
	})

	t.Run("Update改成已存在IP拒绝", func(t *testing.T) {
		mustCreateHost(t, s, "10.1.1.2")
		err := s.Update(id, UpdateHostRequest{PrivateIP: "10.1.1.2"}, testOp, NewRequestID())
		if err == nil || !strings.Contains(err.Error(), "内网IP已存在") {
			t.Fatalf("期望内网IP已存在，实际 %v", err)
		}
	})
}

func TestHostDelete(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	t.Run("未引用删除成功", func(t *testing.T) {
		id := mustCreateHost(t, s, "10.2.1.1")
		if err := s.Delete(id, testOp, NewRequestID()); err != nil {
			t.Fatalf("Delete 失败: %v", err)
		}
		if err := database.DB.First(&models.Host{}, id).Error; err == nil {
			t.Fatal("删除后仍可回查")
		}
		var appCount int64
		database.DB.Model(&models.HostApplication{}).Where("host_id = ?", id).Count(&appCount)
		if appCount != 0 {
			t.Fatalf("申请信息未级联删除: %d", appCount)
		}
	})

	t.Run("被零信任引用拒绝", func(t *testing.T) {
		id := mustCreateHost(t, s, "10.2.1.2")
		zt := models.ZeroTrust{
			ApplyUnit:   "测试单位",
			AccountName: "acct",
			ApplyTime:   "2026-01-02",
		}
		if err := database.DB.Create(&zt).Error; err != nil {
			t.Fatalf("种子零信任失败: %v", err)
		}
		ztTarget := models.ZeroTrustTarget{ZeroTrustID: zt.ID, HostID: id, Ports: "22"}
		if err := database.DB.Create(&ztTarget).Error; err != nil {
			t.Fatalf("种子零信任接入目标失败: %v", err)
		}
		err := s.Delete(id, testOp, NewRequestID())
		if !errors.Is(err, ErrHostReferenced) {
			t.Fatalf("期望 ErrHostReferenced，实际 %v", err)
		}
	})
}

func TestHostFilter(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	mustCreateHost(t, s, "10.3.1.1")
	id2 := mustCreateHost(t, s, "10.3.1.2")
	if err := s.Update(id2, UpdateHostRequest{EnvType: "测试", Status: "已停止"}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	tests := []struct {
		name      string
		req       FilterHostRequest
		wantTotal int64
		wantFirst uint
	}{
		{
			// T-072：列表倒序（最新创建在前）
			name:      "无筛选全量",
			req:       FilterHostRequest{},
			wantTotal: 2,
			wantFirst: id2,
		},
		{
			name:      "按环境过滤",
			req:       FilterHostRequest{EnvType: "测试"},
			wantTotal: 1,
			wantFirst: id2,
		},
		{
			name:      "按状态过滤",
			req:       FilterHostRequest{Status: "已停止"},
			wantTotal: 1,
			wantFirst: id2,
		},
		{
			name:      "分页默认值纠正",
			req:       FilterHostRequest{Page: 0, PageSize: 0},
			wantTotal: 2,
			wantFirst: id2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Filter(tt.req)
			if err != nil {
				t.Fatalf("Filter 失败: %v", err)
			}
			if resp.Total != tt.wantTotal {
				t.Fatalf("total=%d，期望 %d", resp.Total, tt.wantTotal)
			}
			if len(resp.Hosts) == 0 {
				t.Fatal("结果为空")
			}
			if resp.Hosts[0].ID != tt.wantFirst {
				t.Fatalf("首条 id=%d，期望 %d", resp.Hosts[0].ID, tt.wantFirst)
			}
			if resp.Page < 1 || resp.Size < 1 {
				t.Fatalf("分页参数未纠正: page=%d size=%d", resp.Page, resp.Size)
			}
		})
	}
}

// TestHostFilterOrderDesc T-072：列表按创建时间（id）倒序，最新创建的排最前
func TestHostFilterOrderDesc(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	ids := make([]uint, 0, 3)
	for _, ip := range []string{"10.4.1.1", "10.4.1.2", "10.4.1.3"} {
		ids = append(ids, mustCreateHost(t, s, ip))
	}

	resp, err := s.Filter(FilterHostRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Filter 失败: %v", err)
	}
	if len(resp.Hosts) != 3 {
		t.Fatalf("条数=%d，期望 3", len(resp.Hosts))
	}
	for i, want := range []uint{ids[2], ids[1], ids[0]} {
		if resp.Hosts[i].ID != want {
			t.Fatalf("第 %d 条 id=%d，期望 %d（应为 id DESC）", i, resp.Hosts[i].ID, want)
		}
	}

	// 分页：第二页为最早创建的一条
	page2, err := s.Filter(FilterHostRequest{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("Filter 分页失败: %v", err)
	}
	if len(page2.Hosts) != 1 || page2.Hosts[0].ID != ids[0] {
		t.Fatalf("第二页不符: %+v", page2.Hosts)
	}
}

// TestHostObjectStorageField T-072：对象存储字段（原 disk）创建/更新/查询往返
func TestHostObjectStorageField(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	req := baseHostReq("10.4.2.1")
	req.ObjectStorage = 500
	id, err := s.Create(req, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	host, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID 失败: %v", err)
	}
	if host.ObjectStorage != 500 {
		t.Fatalf("object_storage=%d，期望 500", host.ObjectStorage)
	}

	if err = s.Update(id, UpdateHostRequest{ObjectStorage: 1024}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	host, err = s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID 失败: %v", err)
	}
	if host.ObjectStorage != 1024 {
		t.Fatalf("更新后 object_storage=%d，期望 1024", host.ObjectStorage)
	}
}

// TestHostExternalFlag T-078：云外主机标记的筛选、计数与 CSV 往返
func TestHostExternalFlag(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	k8s := true
	in1, err := s.Create(baseHostReq("10.5.1.1"), testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建云内主机失败: %v", err)
	}
	extReq := baseHostReq("10.5.1.2")
	extReq.IsExternal = &k8s
	ext2, err := s.Create(extReq, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建云外主机失败: %v", err)
	}
	_ = in1

	// 默认（不传 is_external）不过滤
	all, err := s.Filter(FilterHostRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Filter 失败: %v", err)
	}
	if all.Total != 2 || all.InternalCount != 1 || all.ExternalCount != 1 {
		t.Fatalf("全量计数不符: total=%d internal=%d external=%d", all.Total, all.InternalCount, all.ExternalCount)
	}

	onlyInternal, err := s.Filter(FilterHostRequest{Page: 1, PageSize: 10, IsExternal: "0"})
	if err != nil {
		t.Fatalf("Filter 云内失败: %v", err)
	}
	if onlyInternal.Total != 1 || onlyInternal.Hosts[0].ID == ext2 || onlyInternal.Hosts[0].IsExternal {
		t.Fatalf("云内筛选不符: %+v", onlyInternal.Hosts)
	}

	onlyExternal, err := s.Filter(FilterHostRequest{Page: 1, PageSize: 10, IsExternal: "1"})
	if err != nil {
		t.Fatalf("Filter 云外失败: %v", err)
	}
	if onlyExternal.Total != 1 || onlyExternal.Hosts[0].ID != ext2 || !onlyExternal.Hosts[0].IsExternal {
		t.Fatalf("云外筛选不符: %+v", onlyExternal.Hosts)
	}

	// 更新为云内
	notExternal := false
	if err = s.Update(ext2, UpdateHostRequest{IsExternal: &notExternal}, testOp, NewRequestID()); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	onlyExternal, err = s.Filter(FilterHostRequest{Page: 1, PageSize: 10, IsExternal: "1"})
	if err != nil {
		t.Fatalf("Filter 失败: %v", err)
	}
	if onlyExternal.Total != 0 {
		t.Fatalf("更新后云外应为 0，实际 %d", onlyExternal.Total)
	}
}

// TestHostExternalCSV T-078/T-080/T-083：CSV 26 列（末列「是否云外主机」）往返，旧 25 列默认云内
func TestHostExternalCSV(t *testing.T) {
	if last := utils.CSVHeaders[len(utils.CSVHeaders)-1]; last != "是否云外主机" {
		t.Fatalf("CSV 末列表头=%q，期望 是否云外主机", last)
	}
	newTestDB(t)
	s := NewHostService()

	row := []string{
		"华东", "i-078", "ext-host", "10.5.2.1", "否",
		"虚拟机", "CentOS", "2", "X86", "4",
		"40", "100", "生产", "否",
		"运行中", "", "单位", "张三", "", "", "", "", "2026-01-02", "", "600", "是",
	}
	req, err := s.ParseCSVRowToCreateHost(row)
	if err != nil {
		t.Fatalf("解析 26 列失败: %v", err)
	}
	if req.IsExternal == nil || !*req.IsExternal {
		t.Fatal("26 列末列「是」应解析为云外")
	}
	if req.ObjectStorage != 600 {
		t.Fatalf("对象存储应为 600，实际 %d", req.ObjectStorage)
	}
	if req.Tags != "" || req.ApplyUnit != "单位" || req.Applicant != "张三" || req.ApplyTime != "2026-01-02" || req.Remark != "" {
		t.Fatalf("列位移后申请字段不符: %+v", req)
	}

	// 旧 25 列：无云外列 → 默认云内
	req, err = s.ParseCSVRowToCreateHost(row[:25])
	if err != nil {
		t.Fatalf("解析 25 列失败: %v", err)
	}
	if req.IsExternal == nil || *req.IsExternal {
		t.Fatal("旧 25 列应默认云内")
	}

	// 导出列与表头等长（先落库一台）
	createReq := baseHostReq("10.5.3.1")
	extTrue := true
	createReq.IsExternal = &extTrue
	if _, err = s.Create(createReq, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建主机失败: %v", err)
	}
	hosts, err := s.Filter(FilterHostRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Filter 失败: %v", err)
	}
	rows := s.ExportToCSVRows(hosts.Hosts)
	if len(rows) != 1 {
		t.Fatalf("导出应 1 行，实际 %d", len(rows))
	}
	if len(rows[0]) != len(utils.CSVHeaders) {
		t.Fatalf("导出行 %d 列，期望 %d 列", len(rows[0]), len(utils.CSVHeaders))
	}
}

// TestBusinessStatsExcludeExternal T-078：业务统计排除云外主机
func TestBusinessStatsExcludeExternal(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	for _, ip := range []string{"10.5.4.1", "10.5.4.2"} {
		req := baseHostReq(ip)
		req.Applicant = "统计人"
		req.Project = "统计项目"
		if _, err := s.Create(req, testOp, NewRequestID()); err != nil {
			t.Fatalf("创建主机失败: %v", err)
		}
	}
	extReq := baseHostReq("10.5.4.3")
	extReq.Applicant = "统计人"
	extReq.Project = "统计项目"
	extTrue := true
	extReq.IsExternal = &extTrue
	if _, err := s.Create(extReq, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建云外主机失败: %v", err)
	}

	stats, err := NewStatsService().GetBusinessStats()
	if err != nil {
		t.Fatalf("业务统计失败: %v", err)
	}
	if stats.Summary.HostCount != 2 {
		t.Fatalf("业务统计应排除云外主机（期望 2），实际 %d", stats.Summary.HostCount)
	}
}

// TestBusinessStatsExcludePhysicalMachine T-079：物理机与裸金属一样不计入 CPU/内存/存储汇总
func TestBusinessStatsExcludePhysicalMachine(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	vmReq := baseHostReq("10.6.1.1")
	vmReq.CPU = 4
	vmReq.Memory = 8
	vmReq.SystemDisk = 40
	vmReq.Applicant = "物理统计"
	if _, err := s.Create(vmReq, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建虚拟机失败: %v", err)
	}
	for i, ip := range []string{"10.6.1.2", "10.6.1.3"} {
		req := baseHostReq(ip)
		req.CPU = 16
		req.Memory = 64
		req.Applicant = "物理统计"
		if i == 0 {
			req.AssetType = "裸金属服务器"
		} else {
			req.AssetType = "物理机"
		}
		if _, err := s.Create(req, testOp, NewRequestID()); err != nil {
			t.Fatalf("创建物理机/裸金属失败: %v", err)
		}
	}

	stats, err := NewStatsService().GetBusinessStats()
	if err != nil {
		t.Fatalf("业务统计失败: %v", err)
	}
	if stats.Summary.HostCount != 3 {
		t.Fatalf("主机台数应含物理机（期望 3），实际 %d", stats.Summary.HostCount)
	}
	for _, g := range stats.ByProject {
		if g.Key != "物理统计" {
			continue
		}
		if g.CPU != 4 || g.Memory != 8 || g.Storage != 40 {
			t.Fatalf("物理机/裸金属不应计入汇总，实际 cpu=%d mem=%d storage=%d", g.CPU, g.Memory, g.Storage)
		}
	}
}

// TestXLSXExternalFlagRoundtrip T-078：xlsx 备份「是否云外主机」列往返
func TestXLSXExternalFlagRoundtrip(t *testing.T) {
	newTestDB(t)
	s := NewHostService()
	req := baseHostReq("10.5.5.1")
	extTrue := true
	req.IsExternal = &extTrue
	if _, err := s.Create(req, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建云外主机失败: %v", err)
	}
	data, err := NewDataExchangeService().Export()
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}

	openTestDB(t, t.Name()+"_target")
	report, err := NewDataExchangeService().Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("导入失败: %v", err)
	}
	if !report.Committed || len(report.Errors) != 0 {
		t.Fatalf("导入异常: committed=%v errors=%+v", report.Committed, report.Errors)
	}
	var host models.Host
	if err = database.DB.Where("private_ip = ?", "10.5.5.1").First(&host).Error; err != nil {
		t.Fatalf("目标库缺主机: %v", err)
	}
	if !host.IsExternal {
		t.Fatal("xlsx 往返应保留云外标记")
	}
}
