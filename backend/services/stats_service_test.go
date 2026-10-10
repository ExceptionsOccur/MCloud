package services

import (
	"net"
	"testing"

	"mcloud/database"
	"mcloud/models"
)

// deadIP TEST-NET-1（192.0.2.0/24，RFC 5737 保留不可达），探测必然不通
const deadIP = "192.0.2.1"

func TestProbeTCP(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听本地端口失败: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port

	if !probeTCP("127.0.0.1", port) {
		t.Fatal("开放端口应可达")
	}

	if err := l.Close(); err != nil {
		t.Fatalf("关闭监听失败: %v", err)
	}
	if probeTCP("127.0.0.1", port) {
		t.Fatal("已关闭端口应不可达")
	}
	if probeTCP(deadIP, 22) {
		t.Fatal("不可达地址应探测失败")
	}
}

// TestPingICMPLoopback 环回 ICMP 应通（环境禁 ICMP socket 时跳过，TCP 兜底由 Probe 链路覆盖）
func TestPingICMPLoopback(t *testing.T) {
	if !pingICMP("127.0.0.1") {
		t.Skip("本环境无可用 ICMP socket（udp4/ip4:icmp 均被拒），跳过回环探测")
	}
}

// TestPingICMPDead 不可达地址在超时窗口内返回 false
func TestPingICMPDead(t *testing.T) {
	if pingICMP(deadIP) {
		t.Fatal("TEST-NET-1 不应探测为通")
	}
}

// TestProbeStateMachineDead 探测不通时颜色状态机与 BUSINESS_LOGIC「结果处理逻辑」一致
func TestProbeStateMachineDead(t *testing.T) {
	newTestDB(t)
	s := NewStatsService()

	// 无记录 + 当前绿 → 绿（保持）
	resp, err := s.Probe(ProbeRequest{IP: deadIP, Color: "green"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "green" {
		t.Fatalf("无记录+绿 应保持绿，实际 %q", resp.Color)
	}

	// 无记录 + 当前非绿 → 黄
	resp, err = s.Probe(ProbeRequest{IP: deadIP, Color: "red"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "yellow" {
		t.Fatalf("无记录+非绿 应为黄，实际 %q", resp.Color)
	}

	// 有记录 + 有数据（cpu>0）→ 红（保持，兼容关机主机）
	host := models.Host{PrivateIP: deadIP, CPU: 4, Name: "dead-with-data"}
	if e := database.DB.Create(&host).Error; e != nil {
		t.Fatalf("种子主机失败: %v", e)
	}
	resp, err = s.Probe(ProbeRequest{IP: deadIP, Color: "green"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "red" {
		t.Fatalf("有记录+有数据 应保持红，实际 %q", resp.Color)
	}

	// 有记录 + 空数据（cpu=0）→ 黄
	host.CPU = 0
	if e := database.DB.Save(&host).Error; e != nil {
		t.Fatalf("更新主机失败: %v", e)
	}
	resp, err = s.Probe(ProbeRequest{IP: deadIP, Color: "red"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "yellow" {
		t.Fatalf("有记录+空数据 应为黄，实际 %q", resp.Color)
	}
}

// TestProbeStateMachineAlive 探测通时颜色状态机（环回 ICMP 可用时执行）
func TestProbeStateMachineAlive(t *testing.T) {
	if !pingICMP("127.0.0.1") {
		t.Skip("本环境无可用 ICMP socket，跳过 alive 状态机")
	}
	newTestDB(t)
	s := NewStatsService()

	// 无记录 + alive → 黄（不新增记录，仅会话内标记）
	resp, err := s.Probe(ProbeRequest{IP: "127.0.0.1", Color: "green"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "yellow" {
		t.Fatalf("无记录+alive 应为黄，实际 %q", resp.Color)
	}
	var count int64
	database.DB.Model(&models.Host{}).Where("private_ip = ?", "127.0.0.1").Count(&count)
	if count != 0 {
		t.Fatalf("探测不应写入 hosts 表，实际 %d 条", count)
	}

	// 有记录 + 有数据 + alive → 红
	host := models.Host{PrivateIP: "127.0.0.1", CPU: 2, Name: "loopback"}
	if e := database.DB.Create(&host).Error; e != nil {
		t.Fatalf("种子主机失败: %v", e)
	}
	resp, err = s.Probe(ProbeRequest{IP: "127.0.0.1", Color: "yellow"})
	if err != nil {
		t.Fatalf("Probe 失败: %v", err)
	}
	if resp.Color != "red" {
		t.Fatalf("有记录+有数据+alive 应为红，实际 %q", resp.Color)
	}
}
