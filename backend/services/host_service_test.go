package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"mcloud/database"
	"mcloud/models"
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
			ApplyTime:   time.Now(),
		}
		if err := database.DB.Create(&zt).Error; err != nil {
			t.Fatalf("种子零信任失败: %v", err)
		}
		ztTarget := models.ZeroTrustTarget{ZeroTrustID: zt.ID, HostID: id, Port: 22}
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

	id1 := mustCreateHost(t, s, "10.3.1.1")
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
			name:      "无筛选全量",
			req:       FilterHostRequest{},
			wantTotal: 2,
			wantFirst: id1,
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
			wantFirst: id1,
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
