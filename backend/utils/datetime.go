package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 日期/时间统一口径（T-040）：
//   - 展示/导出统一 YYYY-MM-DD HH:mm:ss（DateTimeLayout）
//   - 主机申请时间严格 YYYY-MM-DD（DateOnlyLayout），空值合法
//   - JSON 时间字段以 RFC3339 带时区为主，兼容无时区格式（按服务器本地时区解释）
const (
	// DateOnlyLayout YYYY-MM-DD（主机申请时间）
	DateOnlyLayout = "2006-01-02"
	// DateTimeLayout YYYY-MM-DD HH:mm:ss（展示/导出）
	DateTimeLayout = "2006-01-02 15:04:05"
	// DateTimeISOLayout YYYY-MM-DDTHH:mm:ss（无时区 JSON 入参）
	DateTimeISOLayout = "2006-01-02T15:04:05"
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

// ParseFlexibleTime 解析时间入参：RFC3339 带时区优先，
// 兼容无时区 YYYY-MM-DDTHH:mm:ss / YYYY-MM-DD HH:mm:ss（按服务器本地时区解释）
func ParseFlexibleTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("时间不能为空")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	for _, layout := range []string{DateTimeISOLayout, DateTimeLayout} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf(
		"时间格式不正确 %q，应为 RFC3339（如 2026-10-08T10:00:00+08:00）或无时区（如 2026-10-08T10:00:00）", s)
}

// FlexibleTime 可兼容无时区格式的 time.Time JSON 类型（零信任申请时间入参）
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON 解析规则见 ParseFlexibleTime
func (f *FlexibleTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("申请时间必须是时间字符串")
	}
	t, err := ParseFlexibleTime(s)
	if err != nil {
		return err
	}
	f.Time = t
	return nil
}
