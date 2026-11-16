#!/usr/bin/env bash
# 四层门禁：L1 格式与静态检查 / L2 测试与竞态 / L3 基准模糊覆盖 / L4 可复现构建
# 用法：保证 go 在 PATH 中；若工具链位于自定义目录，先 export GOROOT 并把 $GOROOT/bin 加入 PATH。
set -euo pipefail
cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
  echo "go 不在 PATH；请先 export GOROOT 并把 \$GOROOT/bin 加入 PATH" >&2
  exit 1
fi
export GOCACHE="${GOCACHE:-$(go env GOCACHE)}"
export GOPROXY=off
export GOFLAGS=-mod=mod

echo "== L1 格式与静态检查 =="
test -z "$(gofmt -l .)" || { echo "以下文件未格式化："; gofmt -l .; exit 1; }
go vet ./...

echo "== L2 测试、竞态与顺序扰动 =="
go test -race -shuffle=on -count=1 ./...

echo "== L3 基准、覆盖与模糊 =="
go test -bench=. -benchmem ./internal/core
go test -fuzz=FuzzParseSpec -fuzztime=5s ./internal/core
go test -covermode=atomic -coverprofile=cover.out ./...
go tool cover -func=cover.out

echo "== L4 可复现构建与产物元数据 =="
go mod tidy
go mod verify
go build -trimpath -ldflags "-X main.version=release" -o bin/app ./cmd/app
go version -m bin/app

echo "ALL_GATES_PASSED"
