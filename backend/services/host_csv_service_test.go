package services

import (
	"strings"
	"testing"
)

// TestExportCSVFiltered 导出行数与列表筛选 total 一致（T-008 验收）
func TestExportCSVFiltered(t *testing.T) {
	newTestDB(t)
	s := NewHostService()

	mustCreateHost(t, s, "10.4.1.1")
	id2 := mustCreateHost(t, s, "10.4.1.2")
	if err := s.Update(id2, UpdateHostRequest{EnvType: "测试", Status: "已停止"}, testOp, NewRequestID()); err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	tests := []struct {
		name string
		req  FilterHostRequest
	}{
		{name: "无筛选全量", req: FilterHostRequest{}},
		{name: "按环境筛选", req: FilterHostRequest{EnvType: "测试"}},
		{name: "按状态筛选", req: FilterHostRequest{Status: "运行中"}},
		{name: "组合筛选", req: FilterHostRequest{EnvType: "生产", Status: "运行中"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Filter(tt.req)
			if err != nil {
				t.Fatalf("Filter 失败: %v", err)
			}
			csvOut, err := s.ExportCSV(tt.req)
			if err != nil {
				t.Fatalf("ExportCSV 失败: %v", err)
			}
			if !strings.HasPrefix(csvOut, "\uFEFF") {
				t.Fatal("导出缺少 BOM")
			}
			content := strings.TrimPrefix(csvOut, "\uFEFF")
			lines := strings.Split(strings.TrimSpace(content), "\n")
			got := len(lines) - 1 // 去表头
			if int64(got) != resp.Total {
				t.Fatalf("导出行数 %d != 列表 total %d", got, resp.Total)
			}
		})
	}
}
