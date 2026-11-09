# Repository guidance

- Work on Linux with Go 1.27. The netlink reader is Linux-only. Use `make help` for commands; plain `make` displays help. `make check` is non-mutating and works with uncommitted changes.
- Tests use temporary config files and injected collection functions; no root privileges or local `config.yml` are required. Run `make test-race` for collector/lifecycle changes. The netlink decoder has a fuzz target in `pkg/exporter/ipset_linux_test.go`.
- Preserve `ipset_count{set="..."}` compatibility. Collections publish one immutable snapshot; failed collections remove count series and expose `ipset_up=0`, while successful empty collections expose `ipset_up=1`.
- Keep one polling writer per collector. Each exporter owns its Prometheus registry. HTTP scrapes read cached snapshots and perform no netlink I/O.
- Use header-only LIST requests with `IPSET_FLAG_LIST_HEADER`; the high-level library dumps every member and mishandles multipart counts. Counts are kernel-reported and timed-out entries may remain until kernel garbage collection. Keep raw attribute parsing bounds-checked, including alignment padding.
- Live checks need `CAP_NET_ADMIN` in the namespace containing the sets. Use an isolated network namespace for fixtures. Reapply binary capabilities after rebuilding.
- `config.yml` is ignored local state; change `config.yml.example` for shared defaults. Config parsing is independent of CLI flags and rejects unknown YAML fields. CLI precedence is `-config`, `CONFIG_FILE`, then `config.yml`.
- The scratch runtime intentionally uses UID 0 with all capabilities dropped except NET_ADMIN in the deployment examples. Keep the runtime shell-free and read-only compatible. Docker uses BuildKit cache mounts and forwards TARGETVARIANT to GOARM for ARM v6/v7.
