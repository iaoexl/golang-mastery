#!/usr/bin/env bash
# 项目完整性第「1」步：从一个空目录开始，建文件夹、写第一个文件、跑通、打包发布。
# 它验证的是"照着知识库板块1 能不能真的起个项目"，不是复刻主工程。
# 用法：bash scripts/bootstrap.sh [目标空目录]；缺省落到临时目录。
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
  echo "go 不在 PATH；工具链位于自定义目录时先 export GOROOT 并把 \$GOROOT/bin 加入 PATH" >&2
  exit 1
fi
export GOPROXY=off
export GOFLAGS=-mod=mod
# 跟随 go 自身理解的缓存落点，避免 Git Bash 的 /tmp 被 go.exe 解析成盘符路径。
export GOCACHE="${GOCACHE:-$(go env GOCACHE)}"
export GOPATH="${GOPATH:-$(go env GOPATH)}"

TARGET="${1:-${TMPDIR:-/tmp}/atlas_bootstrap}"
# 删除前的安全护栏：拒绝根目录、家目录、当前目录与仓库根，并要求目标为空或不存在。
case "$TARGET" in
  ""|"/"|"."|".."|"$HOME"|"$HOME/"*) 
    echo "目标目录不安全：$TARGET" >&2; exit 1 ;;
esac
if [ -e "$TARGET" ]; then
  if [ ! -d "$TARGET" ]; then
    echo "目标不是目录：$TARGET" >&2; exit 1
  fi
  if [ -n "$(ls -A "$TARGET")" ]; then
    echo "目标目录非空，拒绝清空：$TARGET" >&2; exit 1
  fi
fi
rm -rf "$TARGET"
mkdir -p "$TARGET"
cd "$TARGET"

echo "== 第1步 空目录与文件夹安排 =="
mkdir -p cmd/app internal/core scripts dist
cat > 目录安排.md <<'EOF'
# 目录安排（起步版）

| 路径 | 解决什么 | 放错的后果 |
| --- | --- | --- |
| `go.mod` | 模块边界与依赖集合 | 退回旧工作区模式，导入路径无法解析 |
| `cmd/app/` | 可执行入口，未来多入口各占一个子目录 | 入口与库混放，包名冲突 |
| `internal/core/` | 只被本模块内导入的实现 | 外部误依赖内部细节 |
| `scripts/` | 门禁、快照、演练脚本 | 命令散落在个人笔记里，换机器不可复跑 |
| `dist/`（不入库） | 打包发布产物 | 二进制混进版本库 |
EOF
echo "  已建 cmd/app、internal/core、scripts、dist，并写下目录安排.md"

echo "== 第2步 第一个文件：go.mod =="
go mod init example.com/hello
test -s go.mod && grep -q '^module example.com/hello' go.mod
echo "  go.mod 模块路径确认"

echo "== 第3步 第二个文件：入口 =="
cat > cmd/app/main.go <<'EOF'
package main

import "fmt"

func main() { fmt.Println("hello from the first file") }
EOF
cat > internal/core/word.go <<'EOF'
package core

// Word 是第一个可测的纯函数，用来证明"库包"与"入口包"是两回事。
func Word() string { return "hello" }
EOF
gofmt -w ./cmd ./internal
test -z "$(gofmt -l .)"
echo "  入口与库包各就位，格式化无残留"

echo "== 第4步 第一次跑通 =="
OUT=$(go run ./cmd/app)
[ "$OUT" = "hello from the first file" ]
go vet ./...
cat > internal/core/word_test.go <<'EOF'
package core

import "testing"

func TestWord(t *testing.T) {
	if got := Word(); got != "hello" {
		t.Fatalf("Word() = %q, want hello", got)
	}
}
EOF
gofmt -w ./internal/core/word_test.go
go test ./...
echo "  跑通、静态检查、最小测试全绿"

echo "== 第5步 打包发布 =="
go build -trimpath -o dist/hello ./cmd/app
go version -m dist/hello | grep -E '^\s+(path|build\s+-trimpath)' | sed 's/^/    /' || true
test -x dist/hello
cp 目录安排.md dist/目录安排.md
tar -czf dist/hello-release.tar.gz -C dist hello 目录安排.md
cat > dist/发布清单.md <<'EOF'
# 发布清单（起步版）

| 项 | 内容 |
| --- | --- |
| 产物 | `hello`，由 `go build -trimpath` 生成 |
| 元数据 | `go version -m` 可读出模块路径与构建参数 |
| 源码边界 | `dist/` 不入库，只交付产物与清单 |
| 复跑 | `go build -trimpath -o dist/hello ./cmd/app` |
EOF
echo "  产物、元数据、发布清单、归档包齐备"

echo "== 判定 =="
echo "空目录起步：通过；文件夹安排：通过；第一个文件可跑：通过；打包发布：通过"
echo "BOOTSTRAP_PASSED"
