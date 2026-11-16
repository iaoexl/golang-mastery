package core

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// ErrUnavailable 代表暂时性失败，是唯一允许重试的错误类别。
var ErrUnavailable = errors.New("dependency unavailable")

// FetchWithRetry 演示四问表里的"超时、重试、幂等"三项如何落到代码上：
// key 由入口生成并全程不变，作为幂等键；总耗时由 ctx 统一兜住；
// 指数退避叠加随机抖动，避免多实例同步重试；确定性失败直接返回，不放大无效流量。
func FetchWithRetry(ctx context.Context, key string, call func(context.Context, string) (int, error)) (int, error) {
	const maxAttempts = 4
	base := 2 * time.Millisecond
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("budget exhausted for %q: %w", key, ctx.Err())
		default:
		}

		v, err := call(ctx, key)
		if err == nil {
			return v, nil
		}
		if !errors.Is(err, ErrUnavailable) {
			return 0, fmt.Errorf("non-retryable for %q: %w", key, err)
		}
		last = err

		wait := base << attempt
		jitter := time.Duration(rand.Int63n(int64(wait/2) + 1))
		timer := time.NewTimer(wait + jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return 0, fmt.Errorf("budget exhausted for %q: %w", key, ctx.Err())
		case <-timer.C:
		}
	}
	return 0, fmt.Errorf("attempts exhausted for %q: %w", key, last)
}
