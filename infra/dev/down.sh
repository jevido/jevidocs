#!/usr/bin/env bash
# Stops every local dev server this repo runs, found by the ports they listen
# on (passed as arguments). The whole process group behind each listener is
# stopped, so `task`, `wails3 dev`, `go run` and their children go down
# together. Later this is also where the Podman dev stack gets torn down.
set -u

self=$(ps -o pgid= -p $$ | tr -d ' ')
stopped=0

listeners() {
  if command -v lsof >/dev/null 2>&1; then
    lsof -tiTCP:"$1" -sTCP:LISTEN 2>/dev/null
  else
    ss -ltnpH "sport = :$1" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2
  fi
}

for port in "$@"; do
  for pid in $(listeners "$port" | sort -u); do
    pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
    # Never signal our own group or init's.
    if [ -n "$pgid" ] && [ "$pgid" -gt 1 ] && [ "$pgid" != "$self" ]; then
      kill -TERM -- "-$pgid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null
    else
      kill -TERM "$pid" 2>/dev/null
    fi
    echo "stopped :$port ($(ps -o comm= -p "$pid" 2>/dev/null || echo "pid $pid"))"
    stopped=1
  done
done

# `wails3 dev` runs Vite in a separate process group, so the dev watcher and
# the app binary survive the port sweep above. Stop the ones started from
# inside this repo; other projects' Wails apps are left alone.
root=$(cd "$(dirname "$0")/../.." && pwd -P)

cwd_of() {
  if [ -e "/proc/$1/cwd" ]; then
    readlink "/proc/$1/cwd"
  else
    lsof -a -p "$1" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p'
  fi
}

for pid in $(pgrep -x wails3; pgrep -x bakery); do
  case "$(cwd_of "$pid")/" in
    "$root"/*)
      echo "stopped $(ps -o args= -p "$pid" 2>/dev/null | cut -c1-60)"
      kill -TERM "$pid" 2>/dev/null
      stopped=1
      ;;
  esac
done

[ "$stopped" = 1 ] || echo "nothing running on: $*"
