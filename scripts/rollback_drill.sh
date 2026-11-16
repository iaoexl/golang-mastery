#!/usr/bin/env bash
# 12_5「故障演练与回滚」的本地实物载体。
# 演完整半程：双版本产物 -> 切到新版本写入请求 -> 回退旧版本 -> 逐条重放比对。
# 判定看两件事：探针恢复（服务可用）与数据无损（同一幂等键重放结果一致）。
# 线上多实例流量切换仍属部署侧，见 0_2 边界声明。
set -euo pipefail
cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
  echo "go 不在 PATH；工具链位于自定义目录时先 export GOROOT 并把 \$GOROOT/bin 加入 PATH" >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "curl 不在 PATH；回滚演练需要 curl 探测探针与写入请求" >&2
  exit 1
fi
export GOPROXY=off
export GOFLAGS=-mod=mod
export GOCACHE="${GOCACHE:-$(go env GOCACHE)}"

WORK="${DRILL_WORK:-/tmp/atlas_rollback}"
PORT="${DRILL_PORT:-18099}"
rm -rf "$WORK"
mkdir -p "$WORK/bin"

PID=""
cleanup() { if [ -n "$PID" ]; then kill "$PID" >/dev/null 2>&1 || true; fi; }
trap cleanup EXIT

wait_health() {
  for _ in $(seq 1 30); do
    code=$(curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/healthz" 2>/dev/null || true)
    if [ "$code" = "200" ]; then return 0; fi
    sleep 0.2
  done
  echo "healthz 未在预期时间内恢复" >&2
  return 1
}

start_version() {
  local v="$1"
  echo "$v" > "$WORK/current"
  "$WORK/bin/app.$v" -addr "127.0.0.1:$PORT" >"$WORK/$v.log" 2>&1 &
  PID=$!
  wait_health
  grep -q "\"version\":\"$v\"" "$WORK/$v.log" || {
    echo "注入版本串未出现在启动日志: $v" >&2; exit 1; }
}

stop_server() {
  if [ -n "$PID" ]; then
    kill "$PID" >/dev/null 2>&1 || true
    wait "$PID" 2>/dev/null || true
    PID=""
  fi
}

echo "== 第1步 构建两版产物，验版本可辨识 =="
go build -trimpath -ldflags "-X main.version=prev" -o "$WORK/bin/app.prev" ./cmd/app
go build -trimpath -ldflags "-X main.version=curr" -o "$WORK/bin/app.curr" ./cmd/app
for v in prev curr; do
  go version -m "$WORK/bin/app.$v" | grep -E '^\s+(path|build\s+-trimpath)' | sed 's/^/    /' || true
  echo "  $v 产物已生成，注入版本稍后由启动日志验证"
done

echo "== 第2步 切到新版本，写入三条请求并落状态文件 =="
start_version curr
: > "$WORK/state"
for i in 1 2 3; do
  spec="a:$((i * 7))"
  code=$(curl -sS -o "$WORK/new.$i" -w '%{http_code}' \
    "http://127.0.0.1:$PORT/sum?key=drill-$i&spec=$spec")
  if [ "$code" != "200" ]; then echo "新版本写入失败 $i: $code" >&2; exit 1; fi
  echo "drill-$i|$spec|$code|$(cat "$WORK/new.$i")" >> "$WORK/state"
done
echo "  写入 3 条，状态文件已落盘"

echo "== 第3步 回滚：停新版本，切回旧版本 =="
stop_server
start_version prev
if [ "$(cat "$WORK/current")" != "prev" ]; then echo "指针未回退" >&2; exit 1; fi
grep -q '"version":"prev"' "$WORK/prev.log" || { echo "旧版本日志缺版本串" >&2; exit 1; }
echo "  指针回退，探针恢复，旧版本启动日志可辨识"

echo "== 第4步 逐条重放，验数据无损 =="
while IFS='|' read -r key spec want body; do
  code=$(curl -sS -o "$WORK/replay" -w '%{http_code}' \
    "http://127.0.0.1:$PORT/sum?key=$key&spec=$spec")
  got="$(cat "$WORK/replay")"
  if [ "$code" != "$want" ] || [ "$got" != "$body" ]; then
    echo "重放不一致: $key code=$code want=$want body=$got want_body=$body" >&2
    exit 1
  fi
  echo "  $key 重放一致（code=$code body=$got）"
done < "$WORK/state"

stop_server
echo "== 判定 =="
echo "探针恢复：通过；版本指针回退：通过；幂等重放结果一致：通过"
echo "ROLLBACK_DRILL_PASSED"
