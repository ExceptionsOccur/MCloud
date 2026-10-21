package services

import (
	"bytes"
	"strings"
	"testing"

	"mcloud/database"
	"mcloud/models"

	"github.com/xuri/excelize/v2"
)

// TestParsePortSet T-073：端口列支持掩码 `*`、逗号列表与 `A-B` 范围（可混用），归一为升序合并区间
func TestParsePortSet(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "单口", raw: "22", want: "22"},
		{name: "列表", raw: "443,22", want: "22,443"},
		{name: "范围", raw: "1000-10005", want: "1000-10005"},
		{name: "范围与相邻单口合并", raw: "80,81-90,91", want: "80-91"},
		{name: "掩码全端口", raw: "*", want: "*"},
		{name: "掩码与其它项混用归一全端口", raw: "22,*", want: "*"},
		{name: "范围含掩码归一", raw: "1000-2000,*", want: "*"},
		{name: "空串视为无端口", raw: "  ", want: ""},
		{name: "横线视为无端口", raw: "-", want: ""},
		{name: "越界报错", raw: "70000", wantErr: true},
		{name: "范围倒序报错", raw: "90-80", wantErr: true},
		{name: "非法掩码写法报错", raw: "**", wantErr: true},
		{name: "空项报错", raw: "22,,443", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spans, err := parsePortSet("tcp端口", tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望报错，实际 %v", formatPortSpans(spans))
				}
				return
			}
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}
			if got := formatPortSpans(spans); got != tt.want {
				t.Fatalf("规范化 %q，期望 %q", got, tt.want)
			}
		})
	}
}

// TestZeroTrustPortColumnsExport T-073：同账户名同内网IP 只占一行，tcp/udp 端口分列并集
func TestZeroTrustPortColumnsExport(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	zs := NewZeroTrustService()
	hostA := makeHost(t, hs, "pc-alpha", "10.70.1.1")
	hostB := makeHost(t, hs, "pc-beta", "10.70.1.2")

	// 同账户名两处主机：A 同时有 tcp/udp，B 仅 tcp
	if _, err := zs.Create(ZeroTrustRequest{
		ApplyUnit: "端口单位", AccountName: "pc-account", ExitLocation: "上海", ApplyTime: "2026-02-01",
		Targets: []ZeroTrustTarget{
			{HostID: hostA, Ports: "22,443", Protocol: "tcp"},
			{HostID: hostA, Ports: "53,1000-10005", Protocol: "udp"},
			{HostID: hostB, Ports: "8080", Protocol: "tcp"},
		},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("种子零信任失败: %v", err)
	}

	rows := zeroTrustSheetRows(t)
	if len(rows) != 2 {
		t.Fatalf("期望 2 行（按账户名+内网IP），实际 %d: %v", len(rows), rows)
	}
	if rows[0][6] != "10.70.1.1" || rows[0][7] != "22,443" || rows[0][8] != "53,1000-10005" {
		t.Fatalf("第一行 IP/tcp/udp 不符: %v", rows[0])
	}
	if rows[1][6] != "10.70.1.2" || rows[1][7] != "8080" || rows[1][8] != "" {
		t.Fatalf("第二行 IP/tcp/udp 不符: %v", rows[1])
	}
}

// TestZeroTrustPortColumnsRoundtrip T-073：导出入 → 空库导入，端口按协议回灌；二次导入幂等
func TestZeroTrustPortColumnsRoundtrip(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	zs := NewZeroTrustService()
	hostA := makeHost(t, hs, "pc-rt-alpha", "10.71.1.1")
	if _, err := zs.Create(ZeroTrustRequest{
		ApplyUnit: "回灌单位", AccountName: "pc-rt", ApplyTime: "2026-02-02",
		Targets: []ZeroTrustTarget{
			{HostID: hostA, Ports: "22,1000-10005", Protocol: "tcp"},
			{HostID: hostA, Ports: "53", Protocol: "udp"},
		},
	}, testOp, NewRequestID()); err != nil {
		t.Fatalf("种子零信任失败: %v", err)
	}
	data, err := NewDataExchangeService().Export()
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}

	openTestDB(t, t.Name()+"_target")
	mustCreateHost(t, NewHostService(), "10.71.1.1")
	importer := NewDataExchangeService()
	report, err := importer.Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("导入失败: %v", err)
	}
	if !report.Committed || len(report.Errors) != 0 {
		t.Fatalf("导入异常: committed=%v errors=%+v", report.Committed, report.Errors)
	}

	var zt models.ZeroTrust
	if err = database.DB.Where("account_name = ?", "pc-rt").First(&zt).Error; err != nil {
		t.Fatalf("目标库缺零信任: %v", err)
	}
	rows := ztTargetRows(t, zt.ID)
	if len(rows) != 2 {
		t.Fatalf("子表期望 2 行（tcp/udp 各一行），实际 %+v", rows)
	}
	if rows[0].Ports != "22,1000-10005" || rows[0].Protocol != "tcp" || rows[1].Ports != "53" || rows[1].Protocol != "udp" {
		t.Fatalf("端口回灌不符: %+v", rows)
	}

	report2, err := importer.Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("二次导入失败: %v", err)
	}
	st := report2.stat("零信任")
	if st == nil || st.Created != 0 || st.Updated != 0 || st.Skipped != st.Rows {
		t.Fatalf("二次导入应全部 skip: %+v", st)
	}
}

// TestZeroTrustPortColumnsImportRules T-073：掩码展开、空端口两侧、非法端口与旧列缺失报错
func TestZeroTrustPortColumnsImportRules(t *testing.T) {
	newTestDB(t)
	mustCreateHost(t, NewHostService(), "10.72.1.1")
	mustCreateHost(t, NewHostService(), "10.72.1.2")

	headers := zeroTrustHeaders()

	// 掩码 + 分号多 IP
	data := buildWorkbookWithZeroTrust(t, headers, map[int][]string{
		2: {"掩码单位", "mask-account", "", "", "", "", "10.72.1.1;10.72.1.2", "*", "53", "sys", "2026-03-01", ""},
	})
	report, err := NewDataExchangeService().Import(data, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("导入失败: %v", err)
	}
	if !report.Committed {
		t.Fatalf("掩码导入未提交: %+v", report.Errors)
	}
	var zt models.ZeroTrust
	if err = database.DB.Where("account_name = ?", "mask-account").First(&zt).Error; err != nil {
		t.Fatalf("缺零信任: %v", err)
	}
	rows := ztTargetRows(t, zt.ID)
	if len(rows) != 4 {
		t.Fatalf("掩码导入期望 4 行（2 主机 × tcp/udp），实际 %+v", rows)
	}
	for _, r := range rows {
		want := "53"
		if r.Protocol == "tcp" {
			want = "1-65535"
		}
		if r.Ports != want {
			t.Fatalf("%s 端口期望 %q，实际 %q", r.Protocol, want, r.Ports)
		}
	}

	// 两侧端口皆空 → 行错误；非法端口 → 行错误（整体回滚）
	bad := buildWorkbookWithZeroTrust(t, headers, map[int][]string{
		2: {"错误单位", "bad-account", "", "", "", "", "10.72.1.1", "", "", "", "2026-03-01", ""},
	})
	report, err = NewDataExchangeService().Import(bad, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("导入调用失败: %v", err)
	}
	if report.Committed || len(report.Errors) == 0 {
		t.Fatalf("两侧端口为空应报行错误: %+v", report)
	}

	badPort := buildWorkbookWithZeroTrust(t, headers, map[int][]string{
		2: {"错误单位", "bad-port", "", "", "", "", "10.72.1.1", "70000", "", "", "2026-03-01", ""},
	})
	report, err = NewDataExchangeService().Import(badPort, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("导入调用失败: %v", err)
	}
	if report.Committed || len(report.Errors) == 0 {
		t.Fatalf("非法端口应报行错误: %+v", report)
	}

	// 旧列头「接入目标」（无新列）→ 结构错误
	oldHeaders := make([]string, 0, len(headers))
	for _, h := range headers {
		if strings.HasPrefix(h, "tcp端口") || strings.HasPrefix(h, "udp端口") {
			continue
		}
		oldHeaders = append(oldHeaders, h)
	}
	oldHeaders = append(oldHeaders, "接入目标")
	legacy := buildWorkbookWithZeroTrust(t, oldHeaders, map[int][]string{})
	if _, err := NewDataExchangeService().Import(legacy, testOp, NewRequestID()); err == nil {
		t.Fatal("旧列头备份应返回结构错误")
	}
}

// zeroTrustHeaders 当前「零信任」sheet 列头
func zeroTrustHeaders() []string {
	for _, spec := range dataExchangeSheets {
		if spec.Name == "零信任" {
			return append([]string{}, spec.Headers...)
		}
	}
	return nil
}

// zeroTrustSheetRows 导出后读取「零信任」sheet 的数据行
func zeroTrustSheetRows(t *testing.T) [][]string {
	t.Helper()
	data, err := NewDataExchangeService().Export()
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("打开导出文件失败: %v", err)
	}
	defer f.Close()
	rows, err := f.GetRows("零信任")
	if err != nil {
		t.Fatalf("读零信任 sheet 失败: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("零信任 sheet 为空")
	}
	return rows[1:]
}

// buildWorkbookWithZeroTrust 生成 8 sheet 工作簿，仅在「零信任」写指定行（行号 → 列值）
func buildWorkbookWithZeroTrust(t *testing.T, ztHeaders []string, rowsByNum map[int][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", dataExchangeSheets[0].Name); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	for _, spec := range dataExchangeSheets[1:] {
		if _, err := f.NewSheet(spec.Name); err != nil {
			t.Fatalf("建表失败: %v", err)
		}
	}
	for _, spec := range dataExchangeSheets {
		if err := writeSheetRows(f, spec.Name, nil); err != nil {
			t.Fatalf("写 sheet %s 失败: %v", spec.Name, err)
		}
	}
	for i, h := range ztHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue("零信任", cell, h); err != nil {
			t.Fatalf("写零信任表头失败: %v", err)
		}
	}
	for rowNum, row := range rowsByNum {
		for i, v := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			if err := f.SetCellValue("零信任", cell, v); err != nil {
				t.Fatalf("写零信任行失败: %v", err)
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("生成工作簿失败: %v", err)
	}
	return buf.Bytes()
}
