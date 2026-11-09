# Repository guidance

## Build and checks

- Work on Linux: collection uses Linux ipset netlink APIs. `go.mod` requires Go 1.26.0; the Docker builder uses Go 1.27.
- Build with `go build -o bin/ipset-exporter ./cmd/ipset-exporter` or `make -B`. Plain `make` can miss edits to `main.go`: its source prerequisites explicitly exclude that filename.
- `make format` formats Go sources. `make check` also requires the entire tracked working tree to match HEAD, so it fails on intentional uncommitted changes, not just formatting errors.
- `make lint` requires an externally installed `golangci-lint`; no version is pinned. The release workflow builds/publishes images but does not run tests or lint.

## Tests

- `go test -count=1 ./...` runs the suite; focus one test with `go test ./pkg/config -run '^TestLoadConfigFile$' -count=1 -v`.
- A fresh checkout lacks `pkg/config/test_config.yml` because `.gitignore` excludes it. `TestLoadConfigFile` needs this real file: its `MockFile` is never injected into `Load`. Create the minimal fixture only if missing:

    ```sh
    test -f pkg/config/test_config.yml || printf 'app:\n  host: localhost\n  port: 8080\n' > pkg/config/test_config.yml
    ```

- Config tests mutate `CONFIG_FILE` and `flag.CommandLine`; keep them serial. `Load()` registers and parses a global `-config` flag on every call, so repeated calls need a fresh flag set. The invalid-YAML test currently accepts any load error and does not actually feed its malformed YAML to the loader.
- Current tests cover configuration only and need no netlink privileges; live collection requires `CAP_NET_ADMIN` and access to the network namespace containing the sets.

## Running and configuration

- Initialize local config with `cp config.yml.example config.yml` if absent. The README's copy command omits the destination. `config.yml` is ignored local state; update `config.yml.example` for shared configuration changes.
- After building, use `sudo setcap cap_net_admin+ep ./bin/ipset-exporter` then `./bin/ipset-exporter`. Reapply capabilities after rebuilding. `make run` currently passes its target name (`run`) to `setcap` and has no build prerequisite.
- Config path precedence is `-config` flag, then `CONFIG_FILE`, then `config.yml` relative to the working directory. YAML fields have no defaults or validation; supply a positive `ipset.update_interval` in seconds to avoid a ticker panic.
- The scratch image contains only the binary: mount the YAML and set `CONFIG_FILE` to its container path. Host-set collection needs both `--net host` and `--cap-add=NET_ADMIN` (see README).

## Runtime wiring

- `cmd/ipset-exporter/main.go` loads config and initializes logging in `init()`, before `main`; adding tests to this package also invokes that startup logic.
- `pkg/exporter` collects once before listening, then on a background ticker. HTTP scrapes expose cached `ipset_count{set="..."}` gauges, not fresh netlink queries. Failed collections and disappeared sets currently leave old series intact.
- Metrics use the global default Prometheus registry and package-global `IPSetList`; registration removes Go/process collectors and ignores registration errors. Isolate registry state when testing multiple exporter instances.
