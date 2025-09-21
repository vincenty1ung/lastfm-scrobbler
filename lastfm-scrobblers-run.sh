#!/bin/zsh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)/shell/script"

case "$1" in
  init)
    sh "$SCRIPT_DIR/build_lastfm-scrobblers_launchctl.sh"
    ;;
  start)
    sh "$SCRIPT_DIR/start_lastfm-scrobblers.sh"
    ;;
  stop)
    sh "$SCRIPT_DIR/stop_lastfm-scrobblers.sh"
    ;;
  *)
    echo "用法: $0 {init|start|stop}"
    exit 1
    ;;
esac