#!/usr/bin/env bash
# 环境快照卡用：换机器时先跑本脚本，把原始输出贴进 1_3 的"观察"字段，记录的是差异而非结论。
set -euo pipefail
cd "$(dirname "$0")/.."
if ! command -v go >/dev/null 2>&1; then
  echo "go 不在 PATH；工具链位于自定义目录时先 export GOROOT 并把 \$GOROOT/bin 加入 PATH" >&2
  exit 1
fi
echo "--- go version ---";        go version
echo "--- go env 落点 ---";       go env GOROOT GOMODCACHE GOPATH GOPROXY GOOS GOARCH
echo "--- go run --help（输出样本）---"; go run ./cmd/app --help 2>&1 | head -n 6 || true
echo "--- go build ---";          go build -o bin/baseline ./cmd/app && test -x bin/baseline && echo "binary ok"
echo "--- gofmt -l . ---";        gofmt -l .
echo "--- go vet ./... ---";      go vet ./... && echo "vet ok"
echo "--- go test ./... ---";     go test ./...
echo "--- go mod tidy 差异 ---";  go mod tidy && echo "tidy ok"
