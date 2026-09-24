#!/bin/sh
# drain.sh <command...>: run a server, and when the container is asked to stop,
# keep serving for DRAIN_SECONDS (default 5) before passing the signal on.
#
# Why: Coolify removes the old container milliseconds after the new one turns
# healthy, but Traefik only routes to the new one once it has seen that health
# change, and it batches config updates. Serving a few seconds longer covers
# that gap; the server's own graceful shutdown then finishes in-flight
# requests. See infra/deploy/README.md, "Zero-downtime deploys".
#
# STOP_SIGNAL is what the server expects for a graceful stop (TERM for Go,
# QUIT for nginx).
set -u

"$@" &
pid=$!

drain() {
  sleep "${DRAIN_SECONDS:-5}"
  kill -"${STOP_SIGNAL:-TERM}" "$pid" 2>/dev/null
}
trap drain TERM INT QUIT

# `wait` returns early when a trapped signal arrives; wait again for the
# server to finish its graceful shutdown, and exit with its status.
wait "$pid"
status=$?
if kill -0 "$pid" 2>/dev/null; then
  wait "$pid"
  status=$?
fi
exit "$status"
