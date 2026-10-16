package utils

import "testing"

// TestNormalizeDateLenient xlsx 导入宽容日期解析与归一化（T-065）：
// 标准格式原样归一、斜杠/点/中文/紧凑写法归一为 YYYY-MM-DD、空串合法、两位年份与非法值报错
func TestNormalizeDateLenient(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"2026-10-09", "2026-10-09", false},
		{" 2026-10-09 ", "2026-10-09", false},
		{"2026/9/10", "2026-09-10", false},
		{"2026-9-1", "2026-09-01", false},
		{"2026.9.10", "2026-09-10", false},
		{"2026年9月10日", "2026-09-10", false},
		{"20260910", "2026-09-10", false},
		{"26/9/10", "", true},             // 两位年份不猜
		{"2026-10-08 10:30:00", "", true}, // 带时刻不再接受
		{"not-a-date", "", true},
		{"2026-13-01", "", true}, // 非法月份
	}
	for _, tc := range cases {
		got, err := NormalizeDateLenient("申请时间", tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("NormalizeDateLenient(%q) 期望报错，实际归一为 %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeDateLenient(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeDateLenient(%q) 期望 %q 实际 %q", tc.in, tc.want, got)
		}
	}
}
