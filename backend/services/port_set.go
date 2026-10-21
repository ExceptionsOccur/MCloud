package services

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// 端口集合工具（T-073）：数据备份「零信任」sheet 的 tcp端口/udp端口 两列
// 支持掩码 `*`（全端口 1-65535）、逗号列表与 `A-B` 连续范围，三者可混用。

// portMask 全端口掩码（1-65535）
const portMask = "*"

// portSpan 闭区间 [lo, hi]
type portSpan struct {
	lo int
	hi int
}

// parsePortSet 解析端口列文本为规范化区间集合（升序、相邻/重叠合并）。
// 空串或 `-` 视为无端口（返回 nil）；非法 token 返回错误。
func parsePortSet(field, raw string) ([]portSpan, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "-" {
		return nil, nil
	}

	spans := make([]portSpan, 0, 4)
	for _, tok := range strings.Split(raw, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return nil, fmt.Errorf("%s包含空端口项", field)
		}
		if tok == portMask {
			spans = append(spans, portSpan{lo: 1, hi: 65535})
			continue
		}
		if i := strings.IndexByte(tok, '-'); i >= 0 {
			seg := strings.Split(tok, "-")
			if len(seg) != 2 {
				return nil, fmt.Errorf("%s包含不合法端口 %q（范围 1-65535，连续区间写作起始-结束）", field, tok)
			}
			a, errA := strconv.Atoi(strings.TrimSpace(seg[0]))
			b, errB := strconv.Atoi(strings.TrimSpace(seg[1]))
			if errA != nil || errB != nil || a < 1 || b > 65535 || a > b {
				return nil, fmt.Errorf("%s包含不合法端口范围 %q（1-65535 且起始 ≤ 结束）", field, tok)
			}
			spans = append(spans, portSpan{lo: a, hi: b})
			continue
		}
		n, err := strconv.Atoi(tok)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("%s包含不合法端口 %q（1-65535，可用 * 表示全端口）", field, tok)
		}
		spans = append(spans, portSpan{lo: n, hi: n})
	}
	return mergePortSpans(spans), nil
}

// mergePortSpans 排序并合并重叠/相邻区间；全端口覆盖时归一为单个 1-65535
func mergePortSpans(spans []portSpan) []portSpan {
	if len(spans) == 0 {
		return nil
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].lo != spans[j].lo {
			return spans[i].lo < spans[j].lo
		}
		return spans[i].hi < spans[j].hi
	})
	out := make([]portSpan, 0, len(spans))
	out = append(out, spans[0])
	for _, s := range spans[1:] {
		last := &out[len(out)-1]
		if s.lo <= last.hi+1 {
			if s.hi > last.hi {
				last.hi = s.hi
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

// formatPortSpans 区间集合 → 文本；全端口输出 `*`，范围 `A-B`，单口 `A`，以逗号分隔
func formatPortSpans(spans []portSpan) string {
	if len(spans) == 0 {
		return ""
	}
	if len(spans) == 1 && spans[0].lo == 1 && spans[0].hi == 65535 {
		return portMask
	}
	parts := make([]string, len(spans))
	for i, s := range spans {
		if s.lo == s.hi {
			parts[i] = strconv.Itoa(s.lo)
			continue
		}
		parts[i] = strconv.Itoa(s.lo) + "-" + strconv.Itoa(s.hi)
	}
	return strings.Join(parts, ",")
}

// unionPortText 端口列文本并集 → 规范化文本（用于同一主机多协议端口的合并导出）
func unionPortText(field string, raws ...string) (string, error) {
	var spans []portSpan
	for _, raw := range raws {
		s, err := parsePortSet(field, raw)
		if err != nil {
			return "", err
		}
		spans = append(spans, s...)
	}
	return formatPortSpans(mergePortSpans(spans)), nil
}

// portSetToTokens 区间集合 → 端口 token 列表（与 `parsePortExpr` 输出同口径：
// 范围写作 `A-B`、单口写作 `A`），用于写入 zero_trust_targets.ports
func portSetToTokens(spans []portSpan) []string {
	tokens := make([]string, 0, len(spans))
	for _, s := range spans {
		if s.lo == s.hi {
			tokens = append(tokens, strconv.Itoa(s.lo))
			continue
		}
		tokens = append(tokens, strconv.Itoa(s.lo)+"-"+strconv.Itoa(s.hi))
	}
	return tokens
}
