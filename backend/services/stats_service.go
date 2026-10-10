package services

import (
	"fmt"
	"net"
	"os"
	"time"

	"mcloud/database"
	"mcloud/models"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type StatsService struct{}

func NewStatsService() *StatsService {
	return &StatsService{}
}

type IPUsageSubnet struct {
	Subnet   string   `json:"subnet"`
	UsedIPs  []string `json:"used_ips"`
	EmptyIPs []string `json:"empty_ips"`
	TotalIPs int      `json:"total_ips"`
}

func (s *StatsService) GetIPUsage() ([]IPUsageSubnet, error) {
	var hosts []models.Host
	result := database.DB.Select("private_ip", "cpu").Find(&hosts)
	if result.Error != nil {
		return nil, result.Error
	}

	var subnetRows []models.IPSubnet
	if err := database.DB.Order("cidr ASC").Find(&subnetRows).Error; err != nil {
		return nil, err
	}
	subnets := make([]string, 0, len(subnetRows))
	for _, row := range subnetRows {
		subnets = append(subnets, row.CIDR)
	}

	usedMap := make(map[string]map[string]bool)
	emptyMap := make(map[string]map[string]bool)
	ipnets := make(map[string]*net.IPNet)
	for _, s := range subnets {
		usedMap[s] = make(map[string]bool)
		emptyMap[s] = make(map[string]bool)
		if _, ipnet, err := net.ParseCIDR(s); err == nil {
			ipnets[s] = ipnet
		}
	}

	for _, h := range hosts {
		ipStr := h.PrivateIP
		if len(ipStr) == 0 {
			continue
		}
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		for _, subnet := range subnets {
			ipnet, ok := ipnets[subnet]
			if !ok || !ipnet.Contains(ip) {
				continue
			}
			if h.CPU > 0 {
				usedMap[subnet][ipStr] = true
			} else {
				emptyMap[subnet][ipStr] = true
			}
		}
	}

	var result2 []IPUsageSubnet
	for _, subnet := range subnets {
		used := []string{}
		for ip := range usedMap[subnet] {
			used = append(used, ip)
		}
		empty := []string{}
		for ip := range emptyMap[subnet] {
			empty = append(empty, ip)
		}
		result2 = append(result2, IPUsageSubnet{
			Subnet:   subnet,
			UsedIPs:  used,
			EmptyIPs: empty,
			TotalIPs: 256,
		})
	}

	return result2, nil
}

// probeTimeout 单次探测（ICMP/TCP）超时，与前端 100ms 预期一致
const probeTimeout = 100 * time.Millisecond

// listenICMP 依次尝试非特权 ICMP socket（udp4 ping socket，无需 CAP_NET_RAW）
// → ip4:icmp 原始 socket（需 CAP_NET_RAW，容器内 root 默认具备）；均失败由调用方走 TCP 兜底。
func listenICMP() (net.PacketConn, error) {
	if conn, err := icmp.ListenPacket("udp4", "0.0.0.0"); err == nil {
		return conn, nil
	}
	return icmp.ListenPacket("ip4:icmp", "0.0.0.0")
}

// pingICMP 纯 Go ICMP Echo 探测（T-011，替代宿主机 ping 命令，消除平台差异）
func pingICMP(ip string) bool {
	dst, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return false
	}
	conn, err := listenICMP()
	if err != nil {
		return false
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff, // 非特权 socket 下内核会改写 ID，回包按类型 + 来源 IP 过滤
			Seq:  1,
			Data: []byte("mcloud"),
		},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		return false
	}
	if err := conn.SetDeadline(time.Now().Add(probeTimeout)); err != nil {
		return false
	}
	if _, err := conn.WriteTo(b, dst); err != nil {
		return false
	}

	// 在超时窗口内循环读回包，过滤出目标来源的 Echo Reply
	//（原始 socket 可能收到他进程回包，非特权 socket 可能先收到其他消息）
	buf := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			return false
		}
		var reply *icmp.Message
		reply, err = icmp.ParseMessage(1, buf[:n]) // protocol 1 = ICMP (IPv4)
		if err != nil {
			continue
		}
		if reply.Type != ipv4.ICMPTypeEchoReply {
			continue
		}
		if addr, ok := peer.(*net.IPAddr); ok && addr.IP.Equal(dst.IP) {
			return true
		}
	}
}

func probeTCP(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), probeTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type ProbeRequest struct {
	IP    string `json:"ip" binding:"required"`
	Color string `json:"color" binding:"required"`
}

type ProbeResponse struct {
	Color string `json:"color"`
}

func (s *StatsService) Probe(req ProbeRequest) (*ProbeResponse, error) {
	alive := pingICMP(req.IP)
	if !alive {
		alive = probeTCP(req.IP, 22)
	}
	if !alive {
		alive = probeTCP(req.IP, 3389)
	}

	var host models.Host
	result := database.DB.Where("private_ip = ?", req.IP).First(&host)
	hasRecord := result.Error == nil
	hasData := hasRecord && host.CPU > 0

	var newColor string

	if !alive {
		if !hasRecord {
			if req.Color == "green" {
				newColor = "green"
			} else {
				newColor = "yellow"
			}
		} else if hasData {
			newColor = "red"
		} else {
			newColor = "yellow"
		}
	} else {
		// IP 被占用但表中无主机记录时，不再新增空记录，仅返回颜色
		newColor = "yellow"
		if hasData {
			newColor = "red"
		}
	}

	return &ProbeResponse{Color: newColor}, nil
}
