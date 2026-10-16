package services

import (
	"encoding/json"
	"testing"
)

// TestHostChangeHistory 主机新增/修改写入审计且修改含 before/after 字段级 diff（T-066）
func TestHostChangeHistory(t *testing.T) {
	newTestDB(t)
	hostSvc := NewHostService()
	auditSvc := NewAuditService()

	id := mustCreateHost(t, hostSvc, "10.9.9.1")

	if err := hostSvc.Update(id, UpdateHostRequest{Name: "renamed", Status: "已停止"}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	items, err := auditSvc.ListByResource(ResourceHistoryRequest{ResourceType: "host", ResourceID: id})
	if err != nil {
		t.Fatalf("ListByResource 失败: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("期望新增+修改共 2 条，实际 %d", len(items))
	}
	// 倒序：第一条为 update
	upd := items[0]
	if upd.Action != "update" {
		t.Fatalf("期望首条为 update，实际 %s", upd.Action)
	}
	var detail struct {
		Before map[string]any `json:"before"`
		After  map[string]any `json:"after"`
	}
	if err := json.Unmarshal(upd.Detail, &detail); err != nil {
		t.Fatalf("解析 update detail 失败: %v", err)
	}
	if detail.Before["name"] != "host-10.9.9.1" || detail.After["name"] != "renamed" {
		t.Fatalf("期望 name 由 host-10.9.9.1 变为 renamed，实际 %v → %v", detail.Before["name"], detail.After["name"])
	}
	if detail.Before["status"] == detail.After["status"] {
		t.Fatalf("status 应发生变化: %v", detail.After["status"])
	}

	// 新增条目应带 after 快照
	cre := items[1]
	if cre.Action != "create" {
		t.Fatalf("期望次条为 create，实际 %s", cre.Action)
	}
	var createDetail struct {
		After map[string]any `json:"after"`
	}
	if err := json.Unmarshal(cre.Detail, &createDetail); err != nil {
		t.Fatalf("解析 create detail 失败: %v", err)
	}
	if createDetail.After["private_ip"] != "10.9.9.1" {
		t.Fatalf("期望 create after.private_ip=10.9.9.1，实际 %v", createDetail.After["private_ip"])
	}
}

// TestListByResource 按资源维度查询变更记录：精确匹配 resource_type + resource_id 且按 id 倒序
func TestListByResource(t *testing.T) {
	newTestDB(t)
	svc := NewAuditService()

	hostID1 := uint(1)
	hostID2 := uint(2)
	svc.Record(testOp, "create", "host", &hostID1, DetailDiff{After: map[string]any{"name": "a"}}, NewRequestID())
	svc.Record(testOp, "update", "host", &hostID1, DetailDiff{Before: map[string]any{"name": "a"}, After: map[string]any{"name": "b"}}, NewRequestID())
	svc.Record(testOp, "create", "host", &hostID2, DetailDiff{After: map[string]any{"name": "c"}}, NewRequestID())
	svc.Record(testOp, "create", "zero_trust", &hostID1, DetailDiff{After: map[string]any{"account_name": "z"}}, NewRequestID())

	items, err := svc.ListByResource(ResourceHistoryRequest{ResourceType: "host", ResourceID: hostID1})
	if err != nil {
		t.Fatalf("ListByResource 失败: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("期望 host:%d 命中 2 条，实际 %d", hostID1, len(items))
	}
	if items[0].Action != "update" || items[1].Action != "create" {
		t.Fatalf("期望按 id 倒序（update 后 create 前），实际 %s,%s", items[0].Action, items[1].Action)
	}
	for _, it := range items {
		if it.ResourceType != "host" || it.ResourceID == nil || *it.ResourceID != hostID1 {
			t.Fatalf("命中条目资源维度不符: %+v", it)
		}
	}

	zt, err := svc.ListByResource(ResourceHistoryRequest{ResourceType: "zero_trust", ResourceID: hostID1})
	if err != nil {
		t.Fatalf("ListByResource 失败: %v", err)
	}
	if len(zt) != 1 {
		t.Fatalf("期望 zero_trust:%d 命中 1 条，实际 %d", hostID1, len(zt))
	}
}

// TestListByResourceValidation 缺失 resource_type/resource_id 时报错
func TestListByResourceValidation(t *testing.T) {
	newTestDB(t)
	svc := NewAuditService()

	if _, err := svc.ListByResource(ResourceHistoryRequest{ResourceID: 1}); err == nil {
		t.Fatal("缺失 resource_type 应报错")
	}
	if _, err := svc.ListByResource(ResourceHistoryRequest{ResourceType: "host"}); err == nil {
		t.Fatal("缺失 resource_id 应报错")
	}
}
