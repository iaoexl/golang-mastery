package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"example.com/atlas/internal/core"
)

// version 由构建时的 -ldflags "-X main.version=..." 注入。
var version = "dev"

// newHandler 把处理器从 main 里提出来，测试不必真的启动进程就能验证契约。
func newHandler(store core.Storer, logger *slog.Logger) http.Handler {
	return newHandlerWithEnter(store, logger, nil)
}

// newHandlerWithEnter 在 /slow 进入处理时回调 onEnter，供测试确认请求已在途，
// 避免用固定 sleep 等待造成的时序脆弱。
func newHandlerWithEnter(store core.Storer, logger *slog.Logger, onEnter func()) http.Handler {
	mux := http.NewServeMux()
	// 启动检查：进程初始化是否完成；未完成时编排层不把流量切进来。
	mux.HandleFunc("/startupz", func(w http.ResponseWriter, _ *http.Request) {
		if _, ok := store.Get("__ready__"); !ok {
			http.Error(w, "not started", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// 存活：只判本进程是否还活着，不查依赖。
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// 就绪：真实反映依赖状态，失败时摘流量而不是触发重启。
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if _, ok := store.Get("__ready__"); !ok {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		n, err := core.ParseSpec(r.URL.Query().Get("spec"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.Put(r.URL.Query().Get("key"), n); err != nil {
			http.Error(w, "store write failed", http.StatusInternalServerError)
			return
		}
		logger.Info("sum computed", "request_id", r.Header.Get("X-Request-Id"), "value", n)
		// 回显计算值，让回滚重放能比较实际结果而非固定串。
		_, _ = w.Write([]byte(strconv.Itoa(n)))
	})
	// 慢端点只服务于"优雅关闭等待在途请求"的演练。
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		if onEnter != nil {
			onEnter()
		}
		select {
		case <-time.After(30 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	})
	return mux
}

func main() {
	addr := flag.String("addr", "127.0.0.1:0", "listen address")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	store := core.NewMemoryStore()
	srv := &http.Server{
		Addr:              *addr,
		Handler:           newHandler(store, logger),
		ReadHeaderTimeout: 2 * time.Second,
	}

	// 依赖就绪标记：就绪探针据此返回 200。真实工程应由依赖健康检查驱动，
	// 这里用内存存储初始化完成后的一次显式写入表示"依赖已就绪"。
	if err := store.Put("__ready__", 1); err != nil {
		logger.Error("mark ready failed", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		logger.Info("server start", "version", version, "addr", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		logger.Error("serve failed", "err", err)
		os.Exit(1)
	case <-ctx.Done():
	}
	logger.Info("shutdown begin")
	shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}
	logger.Info("shutdown done")
}
