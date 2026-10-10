package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

// Operator 审计操作人快照（controller 从 JWT user_id 解析后下传）
type Operator struct {
	ID   uint
	Name string
}

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

// NewRequestID 生成请求内聚合用 ID（同一 HTTP 请求多条审计共用）
func NewRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b)
}

// GetOperator 按 user_id 取操作人快照（controller 下传前调用）
func GetOperator(userID uint) (Operator, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return Operator{}, errors.New("用户不存在")
	}
	return Operator{ID: user.ID, Name: user.Username}, nil
}

// DetailDiff 前后值 diff；before/after 为 nil 时省略
type DetailDiff struct {
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

// Record 写入一条审计日志；detail 为 nil 时不落 detail 列。
// 失败仅记不阻断业务（审计不改变 CUD 结果语义）。
func (s *AuditService) Record(op Operator, action, resourceType string, resourceID *uint, detail interface{}, requestID string) {
	if requestID == "" {
		requestID = NewRequestID()
	}

	var detailJSON []byte
	if detail != nil {
		if b, err := json.Marshal(detail); err == nil {
			detailJSON = b
		}
	}

	entry := models.AuditLog{
		OperatorID:   op.ID,
		OperatorName: op.Name,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detailJSON,
		RequestID:    requestID,
	}
	// 审计写入失败不回滚业务；生产可换日志库
	_ = database.DB.Create(&entry).Error
}

type ListAuditLogRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	OperatorName string `form:"operator_name"`
	Action       string `form:"action"`
	ResourceType string `form:"resource_type"`
	RequestID    string `form:"request_id"`
	Keyword      string `form:"keyword"`
}

type AuditLogItem struct {
	ID           uint            `json:"id"`
	OperatorID   uint            `json:"operator_id"`
	OperatorName string          `json:"operator_name"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *uint           `json:"resource_id"`
	Detail       json.RawMessage `json:"detail"`
	RequestID    string          `json:"request_id"`
	CreatedAt    string          `json:"created_at"`
}

type ListAuditLogResponse struct {
	Items []AuditLogItem `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}

func (s *AuditService) List(req ListAuditLogRequest) (*ListAuditLogResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	db := database.DB.Model(&models.AuditLog{})
	if v := strings.TrimSpace(req.OperatorName); v != "" {
		db = db.Where("operator_name ILIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(req.Action); v != "" {
		db = db.Where("action = ?", v)
	}
	if v := strings.TrimSpace(req.ResourceType); v != "" {
		db = db.Where("resource_type = ?", v)
	}
	if v := strings.TrimSpace(req.RequestID); v != "" {
		db = db.Where("request_id = ?", v)
	}
	if v := strings.TrimSpace(req.Keyword); v != "" {
		like := "%" + v + "%"
		db = db.Where(`
			operator_name ILIKE ? OR action ILIKE ? OR resource_type ILIKE ? OR request_id ILIKE ?
			OR resource_id::text = ?
		`, like, like, like, like, v)
	}

	var total int64
	db.Count(&total)

	var logs []models.AuditLog
	offset := (req.Page - 1) * req.PageSize
	if err := db.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	items := make([]AuditLogItem, 0, len(logs))
	for _, l := range logs {
		items = append(items, AuditLogItem{
			ID:           l.ID,
			OperatorID:   l.OperatorID,
			OperatorName: l.OperatorName,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceID:   l.ResourceID,
			Detail:       json.RawMessage(l.Detail),
			RequestID:    l.RequestID,
			CreatedAt:    l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ListAuditLogResponse{
		Items: items,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}
