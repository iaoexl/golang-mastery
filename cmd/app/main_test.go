package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/atlas/internal/core"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandlerContract(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		ready bool
		want  int
	}{
		{"liveness", "/healthz", false, http.StatusOK},
		{"startup before init", "/startupz", false, http.StatusServiceUnavailable},
		{"startup after init", "/startupz", true, http.StatusOK},
		{"readiness before dependency", "/readyz", false, http.StatusServiceUnavailable},
		{"readiness after dependency", "/readyz", true, http.StatusOK},
		{"valid sum", "/sum?key=a&spec=a:3", false, http.StatusOK},
		{"bad spec", "/sum?key=a&spec=oops", false, http.StatusBadRequest},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			// 每个子测试用独立 store，避免 -shuffle=on 下共享状态导致的顺序依赖。
			local := core.NewMemoryStore()
			if c.ready {
				_ = local.Put("__ready__", 1)
			}
			localSrv := httptest.NewServer(newHandler(local, testLogger()))
			defer localSrv.Close()
			resp, err := http.Get(localSrv.URL + c.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != c.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, c.want)
			}
		})
	}
}

// httptest.Server.Close 会阻塞到在途请求收敛，正好用来验证
// "停止接新 -> 等待存量 -> 释放" 这条链路不会把请求硬掐断。
func TestGracefulShutdownWaitsInFlight(t *testing.T) {
	entered := make(chan struct{}, 1)
	srv := httptest.NewServer(newHandlerWithEnter(core.NewMemoryStore(), testLogger(), func() {
		entered <- struct{}{}
	}))
	defer srv.Close()

	done := make(chan int, 1)
	go func() {
		resp, err := http.Get(srv.URL + "/slow")
		if err != nil {
			done <- 0
			return
		}
		done <- resp.StatusCode
		resp.Body.Close()
	}()
	// 用服务端回调确认请求真正进入处理器，替代固定 sleep。
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not enter the slow handler in time")
	}
	srv.Close() // 触发优雅关闭并等待收敛
	if code := <-done; code != http.StatusOK {
		t.Fatalf("in-flight request was cut short: status %d", code)
	}
}

func TestVersionDefault(t *testing.T) {
	// 构建时注入 version；未注入时应为源码里的缺省值。
	// 注入值是否生效由 scripts/rollback_drill.sh 读启动日志验证，单测只守缺省值非空。
	if version == "" {
		t.Fatal("version must have a non-empty default")
	}
}
