package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

type ZeroTrustService struct{}

func NewZeroTrustService() *ZeroTrustService {
	return &ZeroTrustService{}
}

type ZeroTrustRequest struct {
	ApplyUnit   string     `json:"apply_unit" binding:"required"`
	AccountName string     `json:"account_name" binding:"required"`
	Contact     string     `json:"contact"`
	HostID      uint       `json:"host_id" binding:"required"`
	Port        int        `json:"port" binding:"required"`
	SystemName  string     `json:"system_name"`
	ApplyTime   *time.Time `json:"apply_time"`
	Remark      string     `json:"remark"`
}

type ZeroTrustItem struct {
	models.ZeroTrust
	HostName  string `json:"host_name"`
	PrivateIP string `json:"private_ip"`
}

// ErrZeroTrustNotFound 零信任台账记录不存在
var ErrZeroTrustNotFound = errors.New("零信任台账记录不存在")

// ErrHostNotFound 主机不存在
var ErrHostNotFound = errors.New("主机不存在")

// ErrHostReferenced 主机已被零信任台账引用，禁止删除
var ErrHostReferenced = errors.New("该主机已被零信任台账引用")

func normalizeZeroTrust(req ZeroTrustRequest) (ZeroTrustRequest, error) {
	req.ApplyUnit = strings.TrimSpace(req.ApplyUnit)
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Contact = strings.TrimSpace(req.Contact)
	req.SystemName = strings.TrimSpace(req.SystemName)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.ApplyUnit == "" {
		return req, errors.New("申请单位不能为空")
	}
	if req.AccountName == "" {
		return req, errors.New("账户名不能为空")
	}
	if req.HostID == 0 {
		return req, errors.New("申请主机不能为空")
	}
	if req.Port < 1 || req.Port > 65535 {
		return req, errors.New("申请端口必须在 1-65535 之间")
	}
	if req.ApplyTime == nil {
		now := time.Now()
		req.ApplyTime = &now
	}
	return req, nil
}

func (s *ZeroTrustService) List(keyword string) ([]ZeroTrustItem, error) {
	db := database.DB.Model(&models.ZeroTrust{}).Preload("Host")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(
			"zero_trusts.apply_unit ILIKE ? OR zero_trusts.account_name ILIKE ? OR zero_trusts.contact ILIKE ? OR zero_trusts.system_name ILIKE ? OR zero_trusts.remark ILIKE ? OR EXISTS (SELECT 1 FROM hosts WHERE hosts.id = zero_trusts.host_id AND (hosts.name ILIKE ? OR hosts.private_ip ILIKE ?))",
			like, like, like, like, like, like, like,
		)
	}

	var records []models.ZeroTrust
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]ZeroTrustItem, 0, len(records))
	for _, r := range records {
		item := ZeroTrustItem{ZeroTrust: r}
		if r.Host != nil {
			item.HostName = r.Host.Name
			item.PrivateIP = r.Host.PrivateIP
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ZeroTrustService) GetByID(id uint) (*models.ZeroTrust, error) {
	var record models.ZeroTrust
	if err := database.DB.Preload("Host").First(&record, id).Error; err != nil {
		return nil, ErrZeroTrustNotFound
	}
	return &record, nil
}

func (s *ZeroTrustService) Create(req ZeroTrustRequest) (uint, error) {
	req, err := normalizeZeroTrust(req)
	if err != nil {
		return 0, err
	}

	var hostCount int64
	database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
	if hostCount == 0 {
		return 0, ErrHostNotFound
	}

	record := models.ZeroTrust{
		ApplyUnit:   req.ApplyUnit,
		AccountName: req.AccountName,
		Contact:     req.Contact,
		HostID:      req.HostID,
		Port:        req.Port,
		SystemName:  req.SystemName,
		ApplyTime:   *req.ApplyTime,
		Remark:      req.Remark,
	}
	if err := database.DB.Create(&record).Error; err != nil {
		return 0, err
	}
	return record.ID, nil
}

func (s *ZeroTrustService) Update(id uint, req ZeroTrustRequest) error {
	req, err := normalizeZeroTrust(req)
	if err != nil {
		return err
	}

	var record models.ZeroTrust
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrZeroTrustNotFound
	}

	var hostCount int64
	database.DB.Model(&models.Host{}).Where("id = ?", req.HostID).Count(&hostCount)
	if hostCount == 0 {
		return ErrHostNotFound
	}

	record.ApplyUnit = req.ApplyUnit
	record.AccountName = req.AccountName
	record.Contact = req.Contact
	record.HostID = req.HostID
	record.Port = req.Port
	record.SystemName = req.SystemName
	record.ApplyTime = *req.ApplyTime
	record.Remark = req.Remark
	return database.DB.Save(&record).Error
}

func (s *ZeroTrustService) Delete(id uint) error {
	var record models.ZeroTrust
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrZeroTrustNotFound
	}
	return database.DB.Delete(&record).Error
}

// HostReferencedByZeroTrust 主机是否被零信任台账引用（供 host_service 删除前校验）
func HostReferencedByZeroTrust(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.ZeroTrust{}).Where("host_id = ?", hostID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询零信任台账引用失败: %w", err)
	}
	return count, nil
}

// BatchCreateText 零信任批量添加：列顺序 申请单位,账户名,联系方式,内网IP,申请端口,系统名称,申请时间,备注
// 主机按内网IP定位；不存在则该行失败；合法行全部插入（无唯一约束，不跳过）
func (s *ZeroTrustService) BatchCreateText(text string) (*BatchCreateResponse, error) {
	resp := &BatchCreateResponse{}
	headerSkipped := false

	for i, raw := range strings.Split(text, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		row, err := utils.ParseCSVLine(line)
		if err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 格式错误(%v)", lineNo, err))
			continue
		}

		if !headerSkipped && len(row) > 0 && strings.TrimSpace(row[0]) == "申请单位" {
			headerSkipped = true
			continue
		}

		if len(row) > 8 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 列数超出 8 列", lineNo))
			continue
		}
		if len(row) < 4 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 至少需要4列(申请单位,账户名,内网IP,申请端口)", lineNo))
			continue
		}
		for len(row) < 8 {
			row = append(row, "")
		}

		privateIP := strings.TrimSpace(row[3])
		var host models.Host
		if hostErr := database.DB.Where("private_ip = ?", privateIP).First(&host).Error; hostErr != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 内网主机不存在(%s)", lineNo, privateIP))
			continue
		}

		port, err := strconv.Atoi(strings.TrimSpace(row[4]))
		if err != nil || port < 1 || port > 65535 {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 申请端口必须在 1-65535 之间", lineNo))
			continue
		}

		req := ZeroTrustRequest{
			ApplyUnit:   strings.TrimSpace(row[0]),
			AccountName: strings.TrimSpace(row[1]),
			Contact:     strings.TrimSpace(row[2]),
			HostID:      host.ID,
			Port:        port,
			SystemName:  strings.TrimSpace(row[5]),
			Remark:      strings.TrimSpace(row[7]),
		}
		if t := strings.TrimSpace(row[6]); t != "" {
			parsed, err := time.ParseInLocation("2006-01-02T15:04:05", t, time.Local)
			if err != nil {
				resp.Errors++
				resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: 申请时间格式应为 YYYY-MM-DDTHH:mm:ss", lineNo))
				continue
			}
			req.ApplyTime = &parsed
		}

		if _, err := s.Create(req); err != nil {
			resp.Errors++
			resp.LineErrors = append(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
		} else {
			resp.Success++
		}
	}

	return resp, nil
}
