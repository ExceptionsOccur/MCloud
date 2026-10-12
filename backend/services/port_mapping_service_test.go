package services

import (
	"errors"
	"strings"
	"testing"

	"mcloud/database"
	"mcloud/models"
)

// TestNormalizePortMappingProtocols 端口协议归一化：缺省补 tcp、大小写归一、
// 数量不一致与非法协议报错（T-055）
func TestNormalizePortMappingProtocols(t *testing.T) {
	cases := []struct {
		name       string
		req        PortMappingRequest
		wantExt    string
		wantIntnl  string
		wantErrSub string
	}{
		{
			name: "缺省协议补齐 tcp",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "80,443", InternalPorts: "8080,8443",
			},
			wantExt: "tcp,tcp", wantIntnl: "tcp,tcp",
		},
		{
			name: "显式协议小写化",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "53,123", InternalPorts: "5353,1234",
				ExternalProtocols: "UDP,udp", InternalProtocols: "udp,TCP",
			},
			wantExt: "udp,udp", wantIntnl: "udp,tcp",
		},
		{
			name: "协议数量与端口不一致",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "80,443", InternalPorts: "8080,8443",
				ExternalProtocols: "tcp",
			},
			wantErrSub: "外网协议数量必须与端口数量一致",
		},
		{
			name: "非法协议值",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "80", InternalPorts: "8080",
				ExternalProtocols: "sctp", InternalProtocols: "tcp",
			},
			wantErrSub: "不合法协议",
		},
		{
			name: "外网与内网端口数量不一致",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "80,443", InternalPorts: "8080",
			},
			wantErrSub: "数量必须一致",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _, _, err := normalizePortMapping(c.req)
			if c.wantErrSub != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("期望错误包含 %q，实际 %v", c.wantErrSub, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("归一化失败: %v", err)
			}
			if got.ExternalProtocols != c.wantExt {
				t.Errorf("外网协议 = %q，期望 %q", got.ExternalProtocols, c.wantExt)
			}
			if got.InternalProtocols != c.wantIntnl {
				t.Errorf("内网协议 = %q，期望 %q", got.InternalProtocols, c.wantIntnl)
			}
		})
	}
}

// TestNormalizePortMappingPortRanges 端口范围表达式：token 规范化（去前导零）、
// 展开后数量配对（范围↔单口混搭）、非法范围报错、协议按 token 等长（T-055）
func TestNormalizePortMappingPortRanges(t *testing.T) {
	cases := []struct {
		name       string
		req        PortMappingRequest
		wantExt    string
		wantIntnl  string
		wantExtP   string
		wantIntnlP string
		wantErrSub string
	}{
		{
			name: "范围与单口混搭按展开数量配对",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1000-1005,80",
				InternalPorts: "1000,1001,1002,1003,1004,1005,8080",
			},
			wantExt:    "1000-1005,80",
			wantIntnl:  "1000,1001,1002,1003,1004,1005,8080",
			wantExtP:   "tcp,tcp",
			wantIntnlP: "tcp,tcp,tcp,tcp,tcp,tcp,tcp",
		},
		{
			name: "范围对范围",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "80-82,53", InternalPorts: "8080-8082,5353",
			},
			wantExt:    "80-82,53",
			wantIntnl:  "8080-8082,5353",
			wantExtP:   "tcp,tcp",
			wantIntnlP: "tcp,tcp",
		},
		{
			name: "范围 token 协议缺省补齐 1 个",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1000-1005", InternalPorts: "1000-1005",
			},
			wantExt:    "1000-1005",
			wantIntnl:  "1000-1005",
			wantExtP:   "tcp",
			wantIntnlP: "tcp",
		},
		{
			name: "协议按 token 与范围并行",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1000-1005,53", InternalPorts: "1000-1005,5353",
				ExternalProtocols: "udp,tcp", InternalProtocols: "UDP,TCP",
			},
			wantExt:    "1000-1005,53",
			wantIntnl:  "1000-1005,5353",
			wantExtP:   "udp,tcp",
			wantIntnlP: "udp,tcp",
		},
		{
			name: "token 规范化去前导零",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "01000-001005,080", InternalPorts: "8080-8086",
			},
			wantExt:    "1000-1005,80",
			wantIntnl:  "8080-8086",
			wantExtP:   "tcp,tcp",
			wantIntnlP: "tcp",
		},
		{
			name: "起始大于结束",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "10005-1000", InternalPorts: "8080",
			},
			wantErrSub: "不合法端口范围",
		},
		{
			name: "展开后数量不一致",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1000-1005", InternalPorts: "8080-8082",
			},
			wantErrSub: "数量必须一致（外网 6 个，内网 3 个）",
		},
		{
			name: "协议数按 token 数而非展开数",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1000-1005,53", InternalPorts: "1000-1005,5353",
				ExternalProtocols: "tcp",
			},
			wantErrSub: "外网协议数量必须与端口数量一致",
		},
		{
			name: "多段范围非法",
			req: PortMappingRequest{
				PublicIP: "203.0.113.10", HostID: 1,
				ExternalPorts: "1-2-3", InternalPorts: "8080",
			},
			wantErrSub: "不合法端口",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _, _, err := normalizePortMapping(c.req)
			if c.wantErrSub != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("期望错误包含 %q，实际 %v", c.wantErrSub, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("归一化失败: %v", err)
			}
			if got.ExternalPorts != c.wantExt {
				t.Errorf("外网端口 = %q，期望 %q", got.ExternalPorts, c.wantExt)
			}
			if got.InternalPorts != c.wantIntnl {
				t.Errorf("内网端口 = %q，期望 %q", got.InternalPorts, c.wantIntnl)
			}
			if got.ExternalProtocols != c.wantExtP {
				t.Errorf("外网协议 = %q，期望 %q", got.ExternalProtocols, c.wantExtP)
			}
			if got.InternalProtocols != c.wantIntnlP {
				t.Errorf("内网协议 = %q，期望 %q", got.InternalProtocols, c.wantIntnlP)
			}
		})
	}
}

// TestPortMappingCreateProtocolDedup 整组判重含外网协议维度：
// 同组重复 40901、同端口不同协议放行、非法协议 40001、范围表达式同样整组判重（T-055）
func TestPortMappingCreateProtocolDedup(t *testing.T) {
	newTestDB(t)
	if err := database.DB.Create(&models.PublicIP{IP: "203.0.113.10", ISP: "电信", ExitLocation: "上海"}).Error; err != nil {
		t.Fatalf("种子公网IP失败: %v", err)
	}
	hostID := mustCreateHost(t, NewHostService(), "192.168.1.10")
	svc := NewPortMappingService()

	base := PortMappingRequest{
		PublicIP: "203.0.113.10", HostID: hostID,
		ExternalPorts: "80,443", InternalPorts: "8080,8443",
	}
	if _, err := svc.Create(base, testOp, NewRequestID()); err != nil {
		t.Fatalf("首条创建失败: %v", err)
	}

	// 同组完全一致（协议缺省同为 tcp）→ 重复
	if _, err := svc.Create(base, testOp, NewRequestID()); !errors.Is(err, ErrPortMappingExists) {
		t.Fatalf("期望 ErrPortMappingExists，实际 %v", err)
	}

	// 同端口不同协议 → 放行（sqlite 测试库 domain 为全量唯一索引、生产为部分索引，
	// 故本行带独立域名规避空串撞索引，不影响判重语义）
	udp := base
	udp.ExternalProtocols = "udp,udp"
	udp.InternalProtocols = "udp,udp"
	udp.Domain = "udp-protocol.test"
	if _, err := svc.Create(udp, testOp, NewRequestID()); err != nil {
		t.Fatalf("同端口不同协议应放行，实际 %v", err)
	}

	// 非法协议 → 40001 路径（service 层普通错误）
	bad := base
	bad.ExternalProtocols = "sctp,sctp"
	bad.InternalProtocols = "tcp,tcp"
	if _, err := svc.Create(bad, testOp, NewRequestID()); err == nil || !strings.Contains(err.Error(), "不合法协议") {
		t.Fatalf("期望协议校验错误，实际 %v", err)
	}

	// 范围表达式整组判重：同范围同协议 → 重复
	rng := PortMappingRequest{
		PublicIP: "203.0.113.10", HostID: hostID,
		ExternalPorts: "1000-10005", InternalPorts: "1000-10005",
		Domain: "range-protocol.test",
	}
	if _, err := svc.Create(rng, testOp, NewRequestID()); err != nil {
		t.Fatalf("范围映射创建失败: %v", err)
	}
	rngDup := rng
	rngDup.Domain = "range-protocol-2.test"
	if _, err := svc.Create(rngDup, testOp, NewRequestID()); !errors.Is(err, ErrPortMappingExists) {
		t.Fatalf("范围同组重复期望 ErrPortMappingExists，实际 %v", err)
	}

	// 存量归一：三条记录协议列均已落库
	var rows []models.PortMapping
	if err := database.DB.Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatalf("查询映射失败: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("期望 3 条映射，实际 %d", len(rows))
	}
	if rows[0].ExternalProtocols != "tcp,tcp" || rows[1].ExternalProtocols != "udp,udp" || rows[2].ExternalProtocols != "tcp" {
		t.Fatalf("协议列落库不符: %q / %q / %q", rows[0].ExternalProtocols, rows[1].ExternalProtocols, rows[2].ExternalProtocols)
	}
	if rows[2].ExternalPorts != "1000-10005" || rows[2].InternalPorts != "1000-10005" {
		t.Fatalf("范围端口落库不符: %q / %q", rows[2].ExternalPorts, rows[2].InternalPorts)
	}
}
