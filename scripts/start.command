#!/bin/bash
set -euo pipefail
umask 077
RESOURCE_DIR="$(cd "$(dirname "$0")" && pwd)"
export GOCRON_DATA_DIR="${GOCRON_DATA_DIR:-$HOME/Library/Application Support/gocron-sqlite}"
mkdir -p "$GOCRON_DATA_DIR"
cd "$GOCRON_DATA_DIR"
LOCK="$GOCRON_DATA_DIR/launcher.lock"
if ! mkdir "$LOCK" 2>/dev/null; then
  if [ -f "$LOCK/pid" ] && kill -0 "$(cat "$LOCK/pid")" 2>/dev/null; then
    open 'http://127.0.0.1:5920'
    exit 0
  fi
  rm -f "$LOCK/pid"
  rmdir "$LOCK"
  mkdir "$LOCK"
fi
printf '%s\n' "$$" > "$LOCK/pid"
WEB_PID=''
NODE_PID=''
cleanup() {
  trap - EXIT INT TERM HUP
  if [ -n "$WEB_PID" ]; then kill -TERM "$WEB_PID" 2>/dev/null || true; wait "$WEB_PID" 2>/dev/null || true; fi
  if [ -n "$NODE_PID" ]; then kill -TERM "$NODE_PID" 2>/dev/null || true; wait "$NODE_PID" 2>/dev/null || true; fi
  rm -f "$LOCK/pid"
  rmdir "$LOCK" 2>/dev/null || true
}
trap cleanup EXIT
trap 'exit 0' INT TERM HUP
for PORT in 5920 5921; do
  if /usr/bin/nc -z 127.0.0.1 "$PORT" 2>/dev/null; then
    echo "端口 $PORT 已被占用，请先停止占用该端口的程序。"
    exit 1
  fi
done
mkdir -p log
"$RESOURCE_DIR/gocron-node" -s 127.0.0.1:5921 >> log/node.log 2>&1 &
NODE_PID=$!
"$RESOURCE_DIR/gocron" web --host 127.0.0.1 -p 5920 >> log/web.log 2>&1 &
WEB_PID=$!
for ATTEMPT in {1..100}; do
  if ! kill -0 "$WEB_PID" 2>/dev/null || ! kill -0 "$NODE_PID" 2>/dev/null; then
    echo "启动失败，请检查 $GOCRON_DATA_DIR/log"
    exit 1
  fi
  if /usr/bin/curl -fsS http://127.0.0.1:5920/api/install/status >/dev/null 2>&1; then break; fi
  sleep 0.1
done
/usr/bin/curl -fsS http://127.0.0.1:5920/api/install/status >/dev/null
echo "GoCron 已启动：http://127.0.0.1:5920"
echo "数据目录：$GOCRON_DATA_DIR"
echo '首次使用请创建管理员。Shell 任务的主机填写 127.0.0.1，端口 5921。'
echo '保持此窗口开启，按 Control+C 安全停止（等待正在执行的任务结束）。'
if [ "${GOCRON_NO_BROWSER:-0}" != 1 ]; then open 'http://127.0.0.1:5920'; fi
while kill -0 "$WEB_PID" 2>/dev/null && kill -0 "$NODE_PID" 2>/dev/null; do sleep 1; done
