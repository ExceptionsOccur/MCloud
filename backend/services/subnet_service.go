package services

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

type SubnetService struct{}

func NewSubnetService() *SubnetService {
	return &SubnetService{}
}

type SubnetRequest struct {
	CIDR string `json:"cidr" binding:"required"`
}

func (s *SubnetService) List() ([]models.IPSubnet, error) {
	var subnets []models.IPSubnet
	if err := database.DB.Order("cidr ASC").Find(&subnets).Error; err != nil {
		return nil, err
	}
	return subnets, nil
}

// normalizeCIDR 校验并规范化 /24 网段，返回网络地址形式的 CIDR
func normalizeCIDR(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("网段不能为空")
	}
	ip, ipnet, err := net.ParseCIDR(input)
	if err != nil {
		return "", errors.New("网段格式错误，示例：172.17.128.0/24")
	}
	if ip.To4() == nil {
		return "", errors.New("仅支持 IPv4 网段")
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 || ones != 24 {
		return "", errors.New("仅支持 /24 掩码")
	}
	network := ip.Mask(ipnet.Mask)
	return fmt.Sprintf("%s/24", network.String()), nil
}

func (s *SubnetService) Create(req SubnetRequest) (uint, error) {
	cidr, err := normalizeCIDR(req.CIDR)
	if err != nil {
		return 0, err
	}

	var count int64
	database.DB.Model(&models.IPSubnet{}).Where("cidr = ?", cidr).Count(&count)
	if count > 0 {
		return 0, errors.New("该网段已存在")
	}

	subnet := models.IPSubnet{CIDR: cidr}
	if err := database.DB.Create(&subnet).Error; err != nil {
		return 0, err
	}
	return subnet.ID, nil
}

func (s *SubnetService) Update(id uint, req SubnetRequest) error {
	cidr, err := normalizeCIDR(req.CIDR)
	if err != nil {
		return err
	}

	var subnet models.IPSubnet
	if err := database.DB.First(&subnet, id).Error; err != nil {
		return errors.New("网段不存在")
	}

	var count int64
	database.DB.Model(&models.IPSubnet{}).Where("cidr = ? AND id != ?", cidr, id).Count(&count)
	if count > 0 {
		return errors.New("该网段已存在")
	}

	subnet.CIDR = cidr
	return database.DB.Save(&subnet).Error
}

func (s *SubnetService) Delete(id uint) error {
	var subnet models.IPSubnet
	if err := database.DB.First(&subnet, id).Error; err != nil {
		return errors.New("网段不存在")
	}
	return database.DB.Delete(&subnet).Error
}
