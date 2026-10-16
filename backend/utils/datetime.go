package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// 日期/时间统一口径（T-040；T-065 收敛为业务申请时间只到日期）：
//   - 业务申请时间（主机 host_applications / 零信任 zero_trusts）严格 YYYY-MM-DD（DateOnlyLayout），
//     系统不记录业务时间中的具体时刻
//   - 审计类时间（created_at/updated_at/审计日志/用户锁定与登录）保留完整时间戳，展示由前端统一格式化
const (
	// DateOnlyLayout YYYY-MM-DD（业务申请时间）
	DateOnlyLayout = "2006-01-02"
)

// ErrInvalidDate 日期字段格式错误哨兵（controller 映射 40001，消息由具体字段给出）
var ErrInvalidDate = errors.New("日期格式错误")

// dateError 携带具体格式提示的日期错误，经 errors.Is 归约为 ErrInvalidDate
type dateError struct{ msg string }

func (e *dateError) Error() string { return e.msg }

func (e *dateError) Is(target error) bool { return target == ErrInvalidDate }

// ValidateDateOnly 校验字段为严格 YYYY-MM-DD（空串合法），错误信息含格式提示
func ValidateDateOnly(field, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.Parse(DateOnlyLayout, s)
	if err != nil || t.Format(DateOnlyLayout) != s {
		return &dateError{msg: fmt.Sprintf("%s格式应为 YYYY-MM-DD（如 2026-10-09），当前值 %q", field, s)}
	}
	return nil
}

// dateLenientLayouts xlsx 导入宽容接受的日期写法（T-065）：标准短横线补零、
// 斜杠/点/中文年月日分隔且不补零（如 2026/9/10）、8 位紧凑 YYYYMMDD；不支持两位年份（歧义不猜）
var dateLenientLayouts = []string{
	"2006-01-02",
	"2006/1/2",
	"2006-1-2",
	"2006.1.2",
	"2006年1月2日",
	"20060102",
}

// NormalizeDateLenient 宽容解析日期并归一化为 YYYY-MM-DD（仅 xlsx 导入路径使用）：
// 空串合法返回空；可解析 → 统一归一化；均不可解析 → 返回与 ValidateDateOnly 同款错误
func NormalizeDateLenient(field, s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	for _, layout := range dateLenientLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.Format(DateOnlyLayout), nil
		}
	}
	return "", &dateError{msg: fmt.Sprintf("%s格式无法识别（示例: 2026-10-09 或 2026/9/10），当前值 %q", field, s)}
}
