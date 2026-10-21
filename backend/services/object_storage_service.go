package services

import (
	"errors"
	"strings"
	"time"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

// ObjectStorageService 对象存储台账（T-088）：外挂存储，不专属某台主机，可被多台主机共享访问
type ObjectStorageService struct{}

func NewObjectStorageService() *ObjectStorageService {
	return &ObjectStorageService{}
}

// ObjectStorageRequest 新增/编辑请求（T-088）
// AccessKey/SecretKey 为访问存储桶的凭证；编辑时 SecretKey 留空表示「保留原值」（前端只回显脱敏值）
type ObjectStorageRequest struct {
	Bucket    string `json:"bucket" binding:"required"`
	Purpose   string `json:"purpose" binding:"required"`
	SizeGB    int    `json:"size_gb"`
	HostIDs   []uint `json:"host_ids"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Remark    string `json:"remark"`
}

var (
	// ErrObjectStorageNotFound 对象存储记录不存在
	ErrObjectStorageNotFound = errors.New("对象存储记录不存在")
	// ErrObjectStorageBucketExists 存储桶名称已被占用
	ErrObjectStorageBucketExists = errors.New("该存储桶名称已存在")
	// ErrHostReferencedByObjectStorage 主机已被对象存储台账引用
	ErrHostReferencedByObjectStorage = errors.New("该主机已被对象存储台账引用")
)

// ObjectStorageItem 列表项：配对主机按内网IP 排序展示；凭证仅回显 AccessKey 与 SecretKey 脱敏值
// auditSnapshot 审计快照：凭证一律脱敏（绝不落明文 SecretKey）
func (r ObjectStorageRequest) auditSnapshot() ObjectStorageRequest {
	req := r
	req.SecretKey = utils.MaskAPIKey(r.SecretKey)
	return req
}

type ObjectStorageItem struct {
	ID            uint                       `json:"id"`
	Bucket        string                     `json:"bucket"`
	Purpose       string                     `json:"purpose"`
	SizeGB        int                        `json:"size_gb"`
	AccessKey     string                     `json:"access_key"`
	SecretKeyMask string                     `json:"secret_key_mask"`
	SecretKeySet  bool                       `json:"secret_key_set"`
	Remark        string                     `json:"remark"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
	Hosts         []models.ObjectStorageHost `json:"hosts"`
}

func normalizeObjectStorage(req ObjectStorageRequest) (ObjectStorageRequest, error) {
	req.Bucket = strings.TrimSpace(req.Bucket)
	req.Purpose = strings.TrimSpace(req.Purpose)
	req.AccessKey = strings.TrimSpace(req.AccessKey)
	req.SecretKey = strings.TrimSpace(req.SecretKey)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Bucket == "" {
		return req, errors.New("存储桶名称不能为空")
	}
	if req.Purpose == "" {
		return req, errors.New("用途不能为空")
	}
	if req.SizeGB < 0 {
		return req, errors.New("分配大小不能为负数")
	}
	req.HostIDs = uniqueUints(req.HostIDs)
	if len(req.HostIDs) == 0 {
		return req, errors.New("请至少选择一台主机")
	}
	return req, nil
}

func uniqueUints(ids []uint) []uint {
	seen := make(map[uint]bool, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// validateObjectStorageHosts 校验所选主机均存在
func validateObjectStorageHosts(ids []uint) error {
	var count int64
	if err := database.DB.Model(&models.Host{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if int(count) != len(ids) {
		return errors.New("所选主机不存在")
	}
	return nil
}

// loadObjectStorageHosts 批量回填配对主机（含主机简要）
func loadObjectStorageHosts(storages []models.ObjectStorage) (map[uint][]models.ObjectStorageHost, error) {
	byStorage := make(map[uint][]models.ObjectStorageHost, len(storages))
	if len(storages) == 0 {
		return byStorage, nil
	}
	ids := make([]uint, len(storages))
	for i, s := range storages {
		ids[i] = s.ID
	}
	var rows []models.ObjectStorageHost
	if err := database.DB.Preload("Host").Where("object_storage_id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		byStorage[r.ObjectStorageID] = append(byStorage[r.ObjectStorageID], r)
	}
	return byStorage, nil
}

func (s *ObjectStorageService) List(keyword string) ([]ObjectStorageItem, error) {
	db := database.DB.Model(&models.ObjectStorage{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("bucket ILIKE ? OR purpose ILIKE ? OR remark ILIKE ?", like, like, like)
	}
	var records []models.ObjectStorage
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	byStorage, err := loadObjectStorageHosts(records)
	if err != nil {
		return nil, err
	}
	items := make([]ObjectStorageItem, 0, len(records))
	for _, r := range records {
		items = append(items, toObjectStorageItem(r, byStorage[r.ID]))
	}
	return items, nil
}

// toObjectStorageItem 组装列表项：SecretKey 解密后仅回显末四位（密文损坏按未设置处理）
func toObjectStorageItem(r models.ObjectStorage, hosts []models.ObjectStorageHost) ObjectStorageItem {
	plain, err := utils.DecryptSecret(r.SecretKeyEnc, config.AppConfig.AISecretKey)
	if err != nil {
		plain = ""
	}
	return ObjectStorageItem{
		ID:            r.ID,
		Bucket:        r.Bucket,
		Purpose:       r.Purpose,
		SizeGB:        r.SizeGB,
		AccessKey:     r.AccessKey,
		SecretKeyMask: utils.MaskAPIKey(plain),
		SecretKeySet:  plain != "",
		Remark:        r.Remark,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Hosts:         hosts,
	}
}

func (s *ObjectStorageService) Create(req ObjectStorageRequest, op Operator, requestID string) (uint, error) {
	req, err := normalizeObjectStorage(req)
	if err != nil {
		return 0, err
	}
	var count int64
	database.DB.Model(&models.ObjectStorage{}).Where("bucket = ?", req.Bucket).Count(&count)
	if count > 0 {
		return 0, ErrObjectStorageBucketExists
	}
	if err = validateObjectStorageHosts(req.HostIDs); err != nil {
		return 0, err
	}

	secretEnc, err := utils.EncryptSecret(req.SecretKey, config.AppConfig.AISecretKey)
	if err != nil {
		return 0, err
	}
	record := models.ObjectStorage{
		Bucket: req.Bucket, Purpose: req.Purpose, SizeGB: req.SizeGB,
		AccessKey: req.AccessKey, SecretKeyEnc: secretEnc, Remark: req.Remark,
	}
	tx := database.DB.Begin()
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	for _, hostID := range req.HostIDs {
		pair := models.ObjectStorageHost{ObjectStorageID: record.ID, HostID: hostID}
		if err := tx.Create(&pair).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	if err := refreshHostsObjectStorageFlag(req.HostIDs...); err != nil {
		return 0, err
	}
	NewAuditService().Record(op, "create", "object_storage", &record.ID, DetailDiff{After: req.auditSnapshot()}, requestID)
	return record.ID, nil
}

func (s *ObjectStorageService) Update(id uint, req ObjectStorageRequest, op Operator, requestID string) error {
	req, err := normalizeObjectStorage(req)
	if err != nil {
		return err
	}
	var record models.ObjectStorage
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrObjectStorageNotFound
	}
	var count int64
	database.DB.Model(&models.ObjectStorage{}).Where("bucket = ? AND id != ?", req.Bucket, id).Count(&count)
	if count > 0 {
		return ErrObjectStorageBucketExists
	}
	if err := validateObjectStorageHosts(req.HostIDs); err != nil {
		return err
	}

	before := ObjectStorageRequest{
		Bucket: record.Bucket, Purpose: record.Purpose, SizeGB: record.SizeGB,
		AccessKey: record.AccessKey, Remark: record.Remark,
	}
	if plain, decErr := utils.DecryptSecret(record.SecretKeyEnc, config.AppConfig.AISecretKey); decErr == nil {
		before.SecretKey = utils.MaskAPIKey(plain)
	}
	var oldPairs []models.ObjectStorageHost
	if err := database.DB.Where("object_storage_id = ?", id).Order("id ASC").Find(&oldPairs).Error; err != nil {
		return err
	}
	oldHostIDs := make([]uint, 0, len(oldPairs))
	for _, p := range oldPairs {
		before.HostIDs = append(before.HostIDs, p.HostID)
		oldHostIDs = append(oldHostIDs, p.HostID)
	}

	// SecretKey 留空表示保留原值（前端只回显脱敏值，无法回传明文）
	beforeAudit := before.auditSnapshot()
	afterAudit := req.auditSnapshot()
	afterSecretEnc := record.SecretKeyEnc
	if req.SecretKey != "" {
		enc, encErr := utils.EncryptSecret(req.SecretKey, config.AppConfig.AISecretKey)
		if encErr != nil {
			return encErr
		}
		afterSecretEnc = enc
	} else {
		afterAudit.SecretKey = before.SecretKey
	}

	tx := database.DB.Begin()
	updates := map[string]interface{}{
		"bucket": req.Bucket, "purpose": req.Purpose, "size_gb": req.SizeGB, "remark": req.Remark,
		"access_key": req.AccessKey, "secret_key_enc": afterSecretEnc,
	}
	if err := tx.Model(&models.ObjectStorage{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("object_storage_id = ?", id).Delete(&models.ObjectStorageHost{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, hostID := range req.HostIDs {
		pair := models.ObjectStorageHost{ObjectStorageID: id, HostID: hostID}
		if err := tx.Create(&pair).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	if err := refreshHostsObjectStorageFlag(append(oldHostIDs, req.HostIDs...)...); err != nil {
		return err
	}
	NewAuditService().Record(op, "update", "object_storage", &id, DetailDiff{Before: beforeAudit, After: afterAudit}, requestID)
	return nil
}

func (s *ObjectStorageService) Delete(id uint, op Operator, requestID string) error {
	var record models.ObjectStorage
	if err := database.DB.First(&record, id).Error; err != nil {
		return ErrObjectStorageNotFound
	}
	var pairs []models.ObjectStorageHost
	if err := database.DB.Where("object_storage_id = ?", id).Order("id ASC").Find(&pairs).Error; err != nil {
		return err
	}
	before := ObjectStorageRequest{
		Bucket: record.Bucket, Purpose: record.Purpose, SizeGB: record.SizeGB,
		AccessKey: record.AccessKey, Remark: record.Remark,
	}
	if plain, decErr := utils.DecryptSecret(record.SecretKeyEnc, config.AppConfig.AISecretKey); decErr == nil {
		before.SecretKey = utils.MaskAPIKey(plain)
	}
	oldHostIDs := make([]uint, 0, len(pairs))
	for _, p := range pairs {
		before.HostIDs = append(before.HostIDs, p.HostID)
		oldHostIDs = append(oldHostIDs, p.HostID)
	}

	tx := database.DB.Begin()
	if err := tx.Where("object_storage_id = ?", id).Delete(&models.ObjectStorageHost{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	if err := refreshHostsObjectStorageFlag(oldHostIDs...); err != nil {
		return err
	}
	NewAuditService().Record(op, "delete", "object_storage", &id, DetailDiff{Before: before.auditSnapshot()}, requestID)
	return nil
}

// refreshHostsObjectStorageFlag 按配对子表重算主机是否使用对象存储（派生字段，同 ip_mapped，T-088）
// 入参主机去重后逐个重算，0 值忽略
func refreshHostsObjectStorageFlag(hostIDs ...uint) error {
	seen := make(map[uint]bool, len(hostIDs))
	for _, id := range hostIDs {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		var count int64
		if err := database.DB.Model(&models.ObjectStorageHost{}).Where("host_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if err := database.DB.Model(&models.Host{}).Where("id = ?", id).
			Update("has_object_storage", count > 0).Error; err != nil {
			return err
		}
	}
	return nil
}

// HostReferencedByObjectStorage 统计主机在对象存储台账中的引用数（删除主机前校验）
func HostReferencedByObjectStorage(hostID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.ObjectStorageHost{}).Where("host_id = ?", hostID).Count(&count).Error
	return count, err
}

// ObjectStoragesByHost 主机详情回填其共享的对象存储台账
func ObjectStoragesByHost(hostID uint) ([]models.ObjectStorage, error) {
	var pairs []models.ObjectStorageHost
	if err := database.DB.Where("host_id = ?", hostID).Find(&pairs).Error; err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return []models.ObjectStorage{}, nil
	}
	ids := make([]uint, len(pairs))
	for i, p := range pairs {
		ids[i] = p.ObjectStorageID
	}
	var records []models.ObjectStorage
	if err := database.DB.Where("id IN ?", ids).Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
