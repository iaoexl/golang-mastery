// Package core 承载可独立测试的业务实现，不依赖 main 包，便于以"库"的形态被门禁验证。
package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrEmpty 表示输入整体为空或只有空白。
	ErrEmpty = errors.New("empty input")
	// ErrMalformed 表示输入结构不合法，例如缺少冒号分隔符。
	ErrMalformed = errors.New("malformed spec")
	// ErrRange 表示数值越出约定区间。
	ErrRange = errors.New("value out of range")
)

// ParseSpec 解析 "key:count" 形态的记录，返回条数与带上下文的错误。
// 错误链用 %w 保留，调用方可用 errors.Is 追溯哨兵；批量场景用 ParseBatch 聚合。
func ParseSpec(s string) (int, error) {
	if strings.TrimSpace(s) == "" {
		return 0, ErrEmpty
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("bad spec %q: %w", s, ErrMalformed)
	}
	n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, fmt.Errorf("count not numeric in %q: %w", s, err)
	}
	if n < 0 || n > 1000 {
		return 0, fmt.Errorf("count %d in %q: %w", n, s, ErrRange)
	}
	return n, nil
}

// ParseBatch 逐行解析，成功项累加，失败项用 errors.Join 一起返回。
// 与"只返回第一个错误"相比，它保留全部诊断信息，且不丢部分结果。
func ParseBatch(lines []string) (int, error) {
	total := 0
	var errs []error
	for i, line := range lines {
		n, err := ParseSpec(line)
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", i+1, err))
			continue
		}
		total += n
	}
	if len(errs) > 0 {
		return total, errors.Join(errs...)
	}
	return total, nil
}
