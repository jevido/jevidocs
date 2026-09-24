# dev

Local development stack, run with `podman compose`. Not set up yet.

`compose.yml` runs local Postgres 18 on `127.0.0.1:4732` (`task db:up`;
`task db:reset` wipes it). `down.sh` stops every local dev server; `task down`
runs it and stops Postgres too.
