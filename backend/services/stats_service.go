package services

import (
	"context"
	"fmt"
	"mcloud/database"
	"mcloud/models"
	"net"
	"os/exec"
	"strings"
	"time"
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

func pingICMP(ip string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", ip)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "bytes from")
}

func probeTCP(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 100*time.Millisecond)
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
