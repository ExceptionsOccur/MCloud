package services

import (
	"errors"
	"strings"
	"testing"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
)

// newObjStorageTestDB 对象存储测试基建：内存库 + 凭证加密密钥（复用 AI_SECRET_KEY 派生密钥）
func newObjStorageTestDB(t *testing.T) {
	t.Helper()
	newTestDB(t)
	prev := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: "test-secret", JWTExpiresHrs: 1, AISecretKey: "test-ai-secret"}
	t.Cleanup(func() { config.AppConfig = prev })
}

func mustCreateObjHost(t *testing.T, s *HostService, ip string) uint {
	t.Helper()
	id, err := s.Create(baseHostReq(ip), testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建主机 %s 失败: %v", ip, err)
	}
	return id
}

// TestObjectStorageCRUD T-088：桶名必填唯一、用途必填、主机至少 1 台、多主机共享往返
func TestObjectStorageCRUD(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	h1 := mustCreateObjHost(t, hostSvc, "10.90.1.1")
	h2 := mustCreateObjHost(t, hostSvc, "10.90.1.2")

	tests := []struct {
		name       string
		req        ObjectStorageRequest
		wantErrSub string
	}{
		{"存储桶名称为空", ObjectStorageRequest{Purpose: "备份", HostIDs: []uint{h1}}, "存储桶名称不能为空"},
		{"用途为空", ObjectStorageRequest{Bucket: "b1", HostIDs: []uint{h1}}, "用途不能为空"},
		{"主机为空", ObjectStorageRequest{Bucket: "b2", Purpose: "备份"}, "请至少选择一台主机"},
		{"主机不存在", ObjectStorageRequest{Bucket: "b3", Purpose: "备份", HostIDs: []uint{9999}}, "所选主机不存在"},
		{"分配大小为负", ObjectStorageRequest{Bucket: "b4", Purpose: "备份", HostIDs: []uint{h1}, SizeGB: -1}, "分配大小不能为负数"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(tt.req, testOp, NewRequestID())
			if err == nil || err.Error() != tt.wantErrSub {
				t.Fatalf("err=%v，期望 %q", err, tt.wantErrSub)
			}
		})
	}

	id, err := svc.Create(ObjectStorageRequest{
		Bucket: "prod-backup", Purpose: "生产备份", SizeGB: 500, HostIDs: []uint{h1, h2, h1}, Remark: "共享",
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	items, err := svc.List("prod")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(items) != 1 || len(items[0].Hosts) != 2 {
		t.Fatalf("多主机配对错误: %+v", items)
	}
	if items[0].SizeGB != 500 || items[0].Purpose != "生产备份" {
		t.Fatalf("字段不符: %+v", items[0])
	}

	// 桶名唯一
	_, err = svc.Create(ObjectStorageRequest{Bucket: "prod-backup", Purpose: "另一个", HostIDs: []uint{h1}}, testOp, NewRequestID())
	if !errors.Is(err, ErrObjectStorageBucketExists) {
		t.Fatalf("重复桶名应报 ErrObjectStorageBucketExists，实际 %v", err)
	}

	// 更新：主机配对重建为仅 h2
	if err = svc.Update(id, ObjectStorageRequest{Bucket: "prod-backup", Purpose: "生产备份2", SizeGB: 800, HostIDs: []uint{h2}}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	items, err = svc.List("")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(items[0].Hosts) != 1 || items[0].Hosts[0].HostID != h2 || items[0].SizeGB != 800 {
		t.Fatalf("更新后配对/字段不符: %+v", items[0])
	}

	// 主机详情回填
	storages, err := ObjectStoragesByHost(h2)
	if err != nil || len(storages) != 1 {
		t.Fatalf("ObjectStoragesByHost(h2) 期望 1 条，实际 %d err=%v", len(storages), err)
	}
	storages, err = ObjectStoragesByHost(h1)
	if err != nil || len(storages) != 0 {
		t.Fatalf("ObjectStoragesByHost(h1) 期望 0 条，实际 %d err=%v", len(storages), err)
	}

	// 删除：配对行级联清理，主机本身保留
	if err = svc.Delete(id, testOp, NewRequestID()); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	var pairCount int64
	database.DB.Model(&models.ObjectStorageHost{}).Count(&pairCount)
	if pairCount != 0 {
		t.Fatalf("删除后配对行应为 0，实际 %d", pairCount)
	}
	if err = svc.Delete(id, testOp, NewRequestID()); !errors.Is(err, ErrObjectStorageNotFound) {
		t.Fatalf("重复删除应报 ErrObjectStorageNotFound，实际 %v", err)
	}
}

// TestObjectStorageXLSXRoundtrip T-088：备份新增「对象存储」「对象存储主机」两 sheet 往返且二次导入幂等
func TestObjectStorageXLSXRoundtrip(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	h1 := mustCreateObjHost(t, hostSvc, "10.91.1.1")
	h2 := mustCreateObjHost(t, hostSvc, "10.91.1.2")
	if _, err := svc.Create(ObjectStorageRequest{Bucket: "share-b", Purpose: "共享盘", SizeGB: 1024, HostIDs: []uint{h1, h2}, Remark: "r"}, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建失败: %v", err)
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

	items, err := NewObjectStorageService().List("")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(items) != 1 || len(items[0].Hosts) != 2 || items[0].SizeGB != 1024 {
		t.Fatalf("回灌结果不符: %+v", items)
	}

	// 二次导入幂等：无新增，配对全部 skip
	report2, err := NewDataExchangeService().Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("二次导入失败: %v", err)
	}
	for _, st := range report2.Sheets {
		if (st.Sheet == "对象存储" || st.Sheet == "对象存储主机") && (st.Created != 0 || st.Updated != 0) {
			t.Fatalf("二次导入 %s 应全 skip，实际 %+v", st.Sheet, st)
		}
	}
}

// TestHostDeleteBlockedByObjectStorage T-088：主机被对象存储引用时删除返回 ErrHostReferencedByObjectStorage
func TestHostDeleteBlockedByObjectStorage(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	hostID := mustCreateObjHost(t, hostSvc, "10.92.1.1")
	if _, err := svc.Create(ObjectStorageRequest{Bucket: "ref-b", Purpose: "备份", SizeGB: 10, HostIDs: []uint{hostID}}, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	err := hostSvc.Delete(hostID, testOp, NewRequestID())
	if !errors.Is(err, ErrHostReferencedByObjectStorage) {
		t.Fatalf("应报 ErrHostReferencedByObjectStorage，实际 %v", err)
	}

	n, err := HostReferencedByObjectStorage(hostID)
	if err != nil || n != 1 {
		t.Fatalf("引用计数应为 1，实际 %d err=%v", n, err)
	}
}

// TestHostObjectStorageAutoFlag T-088：`hosts.has_object_storage` 由存储桶配对自动重算（增/改/删配对）
func TestHostObjectStorageAutoFlag(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	h1 := mustCreateObjHost(t, hostSvc, "10.93.1.1")
	h2 := mustCreateObjHost(t, hostSvc, "10.93.1.2")

	flag := func(id uint) bool {
		t.Helper()
		var h models.Host
		if err := database.DB.First(&h, id).Error; err != nil {
			t.Fatalf("查询主机 %d 失败: %v", id, err)
		}
		return h.HasObjectStorage
	}

	if flag(h1) || flag(h2) {
		t.Fatal("无配对时 has_object_storage 应为 false")
	}

	id, err := svc.Create(ObjectStorageRequest{Bucket: "auto-b", Purpose: "自动判别", SizeGB: 10, HostIDs: []uint{h1, h2}}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if !flag(h1) || !flag(h2) {
		t.Fatal("创建配对后两台主机 has_object_storage 应为 true")
	}

	// 编辑：仅保留 h1 → h2 回落
	if err = svc.Update(id, ObjectStorageRequest{Bucket: "auto-b", Purpose: "自动判别", SizeGB: 10, HostIDs: []uint{h1}}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	if !flag(h1) || flag(h2) {
		t.Fatalf("编辑配对后应仅 h1 为 true，实际 h1=%v h2=%v", flag(h1), flag(h2))
	}

	// 删除存储桶 → 全部回落
	if err = svc.Delete(id, testOp, NewRequestID()); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if flag(h1) || flag(h2) {
		t.Fatal("删除存储桶后 has_object_storage 应回落为 false")
	}
}

// TestObjectStorageCredentials T-088：AccessKey/SecretKey 密文存储、脱敏回显、留空保留、审计不落明文
func TestObjectStorageCredentials(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	h1 := mustCreateObjHost(t, hostSvc, "10.94.1.1")
	id, err := svc.Create(ObjectStorageRequest{
		Bucket: "cred-b", Purpose: "凭证", SizeGB: 20, HostIDs: []uint{h1},
		AccessKey: "AKIA-OBJ-001", SecretKey: "secret-plain-1234",
	}, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	// 密文落库（非明文）
	var row models.ObjectStorage
	if err = database.DB.First(&row, id).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if row.SecretKeyEnc == "" || row.SecretKeyEnc == "secret-plain-1234" {
		t.Fatalf("SecretKey 应为密文，实际 %q", row.SecretKeyEnc)
	}

	items, err := svc.List("cred-b")
	if err != nil || len(items) != 1 {
		t.Fatalf("List 失败: %v", err)
	}
	it := items[0]
	if it.AccessKey != "AKIA-OBJ-001" {
		t.Fatalf("AccessKey 应明文回显，实际 %q", it.AccessKey)
	}
	if it.SecretKeyMask != "****1234" || !it.SecretKeySet {
		t.Fatalf("SecretKey 应末四位脱敏回显，实际 mask=%q set=%v", it.SecretKeyMask, it.SecretKeySet)
	}

	// 审计详情不得含明文 SecretKey
	var logs []models.AuditLog
	if err = database.DB.Where("resource_type = ?", "object_storage").Find(&logs).Error; err != nil {
		t.Fatalf("查询审计失败: %v", err)
	}
	for _, l := range logs {
		if strings.Contains(string(l.Detail), "secret-plain-1234") {
			t.Fatalf("审计日志泄露明文 SecretKey: %s", string(l.Detail))
		}
	}

	// 编辑时 SecretKey 留空 → 保留原密文
	if err = svc.Update(id, ObjectStorageRequest{
		Bucket: "cred-b", Purpose: "凭证2", SizeGB: 20, HostIDs: []uint{h1}, AccessKey: "AKIA-OBJ-002",
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	items, err = svc.List("cred-b")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if items[0].SecretKeyMask != "****1234" || items[0].AccessKey != "AKIA-OBJ-002" {
		t.Fatalf("留空提交应保留原 SecretKey，实际 %+v", items[0])
	}

	// 显式传入新 SecretKey → 替换
	if err = svc.Update(id, ObjectStorageRequest{
		Bucket: "cred-b", Purpose: "凭证2", SizeGB: 20, HostIDs: []uint{h1},
		AccessKey: "AKIA-OBJ-002", SecretKey: "new-secret-9876",
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}
	items, err = svc.List("cred-b")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if items[0].SecretKeyMask != "****9876" {
		t.Fatalf("SecretKey 应更新为 ****9876，实际 %q", items[0].SecretKeyMask)
	}
}

// TestObjectStorageCredentialsXLSX T-088：备份导出明文凭证并可往返恢复
func TestObjectStorageCredentialsXLSX(t *testing.T) {
	newObjStorageTestDB(t)
	hostSvc := NewHostService()
	svc := NewObjectStorageService()

	h1 := mustCreateObjHost(t, hostSvc, "10.95.1.1")
	if _, err := svc.Create(ObjectStorageRequest{
		Bucket: "bak-b", Purpose: "备份凭证", SizeGB: 30, HostIDs: []uint{h1},
		AccessKey: "AKIA-BAK-001", SecretKey: "bak-secret-4321",
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("创建失败: %v", err)
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

	items, err := NewObjectStorageService().List("bak-b")
	if err != nil || len(items) != 1 {
		t.Fatalf("List 失败: %v", err)
	}
	if items[0].AccessKey != "AKIA-BAK-001" || items[0].SecretKeyMask != "****4321" {
		t.Fatalf("凭证往返不符: %+v", items[0])
	}
}
