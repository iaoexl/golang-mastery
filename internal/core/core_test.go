package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// assertInt 用 t.Helper 把失败行号归到调用处，而不是停在这个工具函数内部。
func assertInt(t *testing.T, got, want int, label string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", label, got, want)
	}
}

func TestParseSpecTable(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    int
		wantErr bool
		wantIs  error
	}{
		{"valid", "region-a:42", 42, false, nil},
		{"empty", "", 0, true, ErrEmpty},
		{"no colon", "region-a", 0, true, ErrMalformed},
		{"not numeric", "region-a:abc", 0, true, nil},
		{"negative", "region-a:-1", 0, true, ErrRange},
		{"too large", "region-a:1001", 0, true, ErrRange},
		{"whitespace", "  region-a: 7 ", 7, false, nil},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSpec(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, c.wantErr)
			}
			if c.wantIs != nil && !errors.Is(err, c.wantIs) {
				t.Fatalf("err %v did not wrap %v", err, c.wantIs)
			}
			assertInt(t, got, c.want, "count")
		})
	}
}

func TestParseBatchJoinsErrors(t *testing.T) {
	total, err := ParseBatch([]string{"a:1", "bad", "b:2", "c:9999"})
	assertInt(t, total, 3, "partial total")
	if err == nil {
		t.Fatal("expected joined error")
	}
	if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), "line 4") {
		t.Fatalf("joined error lost detail: %v", err)
	}
	if !errors.Is(err, ErrRange) {
		t.Fatalf("joined error lost sentinel: %v", err)
	}
}

func TestSnapshotIsIndependent(t *testing.T) {
	src := []int{1, 2, 3}
	snap := Snapshot(src)
	snap[0] = 99
	assertInt(t, src[0], 1, "snapshot must not touch source")

	win := Window(src, 0, 2)
	grown := append(win, 7) // cap 被钉死，append 必然另起数组
	grown[0] = 99
	assertInt(t, src[0], 1, "window append must not touch source")
}

func TestStoreAndGenericSum(t *testing.T) {
	var s Storer = NewMemoryStore()
	if err := s.Put("k", 5); err != nil {
		t.Fatal(err)
	}
	if v, ok := s.Get("k"); !ok || v != 5 {
		t.Fatalf("Get = %v, %v", v, ok)
	}
	assertInt(t, Sum([]int{1, 2, 3}), 6, "Sum over int")
	if got := Sum([]float64{1.5, 2.5}); got != 4.0 {
		t.Fatalf("Sum over float64 = %v", got)
	}
}

func TestCollectConverges(t *testing.T) {
	got, err := Collect(context.Background(), []int{1, 2, 3, 4, 5}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("collected %d results", len(got))
	}
	sort.Ints(got)
	want := []int{2, 4, 6, 8, 10}
	for i := range want {
		assertInt(t, got[i], want[i], "doubled value")
	}
}

func TestCollectCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := Collect(ctx, []int{1, 2, 3}, 2)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if len(got) != 0 {
		t.Fatalf("canceled collect returned %d results, want 0", len(got))
	}
}

func TestFetchWithRetryNonRetryable(t *testing.T) {
	calls := 0
	_, err := FetchWithRetry(context.Background(), "k", func(context.Context, string) (int, error) {
		calls++
		return 0, errors.New("bad parameter")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	assertInt(t, calls, 1, "deterministic failure must not be retried")
}

func TestFetchWithRetryEventuallySucceeds(t *testing.T) {
	calls := 0
	v, err := FetchWithRetry(context.Background(), "k2", func(context.Context, string) (int, error) {
		calls++
		if calls < 3 {
			return 0, ErrUnavailable
		}
		return 7, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, v, 7, "value")
	assertInt(t, calls, 3, "attempts")
}

func TestFetchWithRetryRespectsBudget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Millisecond)
	defer cancel()
	_, err := FetchWithRetry(ctx, "k3", func(context.Context, string) (int, error) {
		return 0, ErrUnavailable
	})
	if err == nil {
		t.Fatal("expected budget error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("budget error lost cause: %v", err)
	}
}

// 演示 t.Cleanup 的时机：子测试的清理在子测试结束时执行，
// 外层 Cleanup 要等整个测试函数返回后才跑，所以在断言处只能看到 "inner"。
func TestCleanupRunsAfterSubtest(t *testing.T) {
	var order []string
	t.Cleanup(func() { order = append(order, "outer") })
	t.Run("inner", func(t *testing.T) {
		t.Cleanup(func() { order = append(order, "inner") })
	})
	if len(order) != 1 || order[0] != "inner" {
		t.Fatalf("cleanup order = %v", order)
	}
}

func TestIsolationHelpers(t *testing.T) {
	t.Setenv("ATLAS_MODE", "test")
	if got := os.Getenv("ATLAS_MODE"); got != "test" {
		t.Fatalf("Setenv did not apply: %q", got)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "case.txt"), []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "case.txt")); err != nil {
		t.Fatalf("TempDir file missing: %v", err)
	}
}
