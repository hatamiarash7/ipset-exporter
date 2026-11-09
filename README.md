# Prometheus Exporter for ipset

A lightweight Linux exporter that exposes [ipset](https://ipset.netfilter.org/) entry counts to [Prometheus](https://prometheus.io/).

The exporter reads set headers directly over netlink, without spawning the `ipset` CLI or downloading every member. It collects in the background and serves cached, consistent snapshots, so scraping does not trigger additional kernel queries.

```text
ipset_count{set="my-set-1"} 41
ipset_count{set="my-set-2"} 23
```

## Getting started

You need Linux with kernel ipset support and `CAP_NET_ADMIN` in the network namespace containing the sets. Building from source requires **Go 1.27 or newer**. The container builder uses `golang:1.27-alpine`.

### Run a binary

```bash
git clone https://github.com/hatamiarash7/ipset-exporter.git
cd ipset-exporter
cp config.yml.example config.yml
make build
./bin/ipset-exporter -check-config
sudo setcap cap_net_admin+ep ./bin/ipset-exporter
./bin/ipset-exporter
```

Edit `config.yml` before starting if you need a different bind address or selected sets. The default address is **127.0.0.1:4613**. The `setcap` command is provided by `libcap2-bin` on Debian/Ubuntu or `libcap` on Alpine. Reapply capabilities whenever you rebuild the binary.

```bash
curl http://127.0.0.1:4613/metrics
./bin/ipset-exporter -version
./bin/ipset-exporter -help
```

### Run with Docker

```bash
docker run -d \
  --name ipset-exporter \
  --restart unless-stopped \
  --network host \
  --cap-drop ALL \
  --cap-add NET_ADMIN \
  --security-opt no-new-privileges:true \
  --read-only \
  --mount type=bind,src="$(pwd)/config.yml",dst=/etc/ipset-exporter/config.yml,readonly \
  hatamiarash7/ipset-exporter:latest
```

Host networking is required to observe **host** sets; bridge networking observes the container's namespace instead. Use Docker Engine on Linux for host ipset monitoring. No published-port mapping is needed with host networking.

The image includes the example configuration at `/etc/ipset-exporter/config.yml` and sets `CONFIG_FILE` to that path. The runtime contains only the static binary, configuration, and license. It runs as UID 0 so the explicitly granted `NET_ADMIN` capability works; the examples drop all other capabilities. It needs no writable filesystem or shell.

### Run with Docker Compose

```bash
cp config.yml.example config.yml
docker compose up -d --no-build
docker compose logs -f ipset-exporter
```

The sample [`docker-compose.yml`](docker-compose.yml) uses host networking, a read-only config mount, restricted capabilities, and bounded log rotation. To build your checkout instead of using the published image:

```bash
APP_VERSION=local docker compose build --pull
docker compose up -d --no-build
```

## Configuration

```yaml
app:
  host: 127.0.0.1
  port: 4613
  log_level: info

ipset:
  update_interval: 15
  names:
    - all
```

- `app.host`: listen address; defaults to `127.0.0.1`. Use `0.0.0.0` for remote IPv4 scrapes or `::` for IPv6. Hostnames are also accepted.
- `app.port`: TCP port, from 1 to 65535; defaults to `4613`.
- `app.log_level`: `panic`, `fatal`, `error`, `warn`, `info`, `debug`, or `trace`; defaults to `info`.
- `ipset.update_interval`: positive integer seconds between collections; defaults to `15`.
- `ipset.names`: set names to monitor, or `all` to include every set; defaults to `[all]`. An explicit empty list is rejected. `all` takes precedence when combined with other names.

Omitted fields receive defaults. Unknown fields, invalid values, duplicate keys, malformed YAML, and multiple YAML documents fail startup with an error. Configuration is read once at startup; restart after changing it.

Configuration path precedence:

1. `-config /path/to/config.yml`
2. Nonempty `CONFIG_FILE` environment variable
3. `config.yml` in the working directory

```bash
./bin/ipset-exporter -config /etc/ipset-exporter/config.yml -check-config
CONFIG_FILE=/etc/ipset-exporter/config.yml ./bin/ipset-exporter
```

`-version` and `-help` work without a configuration file or netlink privileges. `-check-config` validates the YAML without starting a listener or querying ipset.

## Metrics and endpoints

| Metric                                                   | Type    | Meaning                                                       |
| -------------------------------------------------------- | ------- | ------------------------------------------------------------- |
| `ipset_count{set="name"}`                                | Gauge   | Kernel-reported number of entries in an existing selected set |
| `ipset_up`                                               | Gauge   | `1` if the latest collection succeeded, otherwise `0`         |
| `ipset_collection_duration_seconds`                      | Gauge   | Duration of the latest collection                             |
| `ipset_collection_errors_total`                          | Counter | Failed collections since startup                              |
| `ipset_last_success_timestamp_seconds`                   | Gauge   | Last successful collection time; `0` before any success       |
| `ipset_exporter_build_info{version,revision,build_date}` | Gauge   | Build metadata, always `1`                                    |

Collection starts immediately and repeats on the configured interval. Counts are replaced atomically after each poll. Deleted or unselected sets disappear; existing empty sets report zero. A failed collection removes all count series, sets `ipset_up` to zero, and is retried at the next interval. A successful collection of zero sets still reports `ipset_up=1`; this indicates collection health, not whether configured names exist.

Counts come from the kernel's set headers. For sets with entry timeouts, expired entries may remain in the reported total until kernel garbage collection removes them.

| Endpoint   | Purpose                                                                           |
| ---------- | --------------------------------------------------------------------------------- |
| `/metrics` | Prometheus metrics; remains available on collection failure                       |
| `/healthz` | HTTP process liveness (`200`)                                                     |
| `/readyz`  | Collection readiness (`200` after success, `503` before success or after failure) |
| `/`        | Link to metrics                                                                   |

Prometheus configuration example:

```yaml
scrape_configs:
  - job_name: ipset
    scrape_interval: 15s
    static_configs:
      - targets: ["127.0.0.1:4613"]
```

Replace the target with the exporter's address if Prometheus runs elsewhere; inside a bridge-networked Prometheus container, `127.0.0.1` refers to that container. Set `app.host` appropriately and restrict access to trusted monitoring clients; the exporter does not implement TLS or authentication.

Alert on Prometheus's `up{job="ipset"} == 0` for scrape failures and `ipset_up == 0` for kernel collection failures. To detect an expected set disappearing, use an `absent(ipset_count{set="my-set"})` rule with suitable instance scoping.

## Build and development

```bash
make                 # Show grouped help
make build           # Always invoke Go's incremental build
make check           # Non-mutating formatting check, go vet, and tests
make test-race       # Race detector (requires CGO and a C compiler)
make lint            # Requires an installed golangci-lint compatible with Go 1.27
make format          # Apply gofmt
```

Tests create their own fixtures and need no root privileges. They cover configuration, CLI flags, cached metric replacement, failures and recovery, concurrent scrapes, header decoding, HTTP routing, listener errors, and shutdown. Fuzz the netlink decoder with:

```bash
go test ./pkg/exporter -run '^$' -fuzz '^FuzzParseIPSetHeader$' -fuzztime=30s
```

GitHub Actions checks source, tests, race detection, and container builds for pull requests. Release builds run checks before publishing multi-architecture images. Go dependencies are recorded in `go.mod` and `go.sum`.

### Version metadata

Make injects version, Git revision, and UTC build date using linker `-X` flags. By default, the version comes from `git describe --tags --always --dirty`; direct `go build` defaults to `dev`.

```bash
make build VERSION=v1.4.0
./bin/ipset-exporter -version
make docker-build VERSION=v1.4.0 IMAGE_TAG=v1.4.0
```

You can override `GO`, `GOFMT`, `GOLANGCI_LINT`, `BIN_DIR`, `CONFIG_FILE`, `VERSION`, `REVISION`, `BUILD_DATE`, `LDFLAGS`, `DOCKER`, `IMAGE`, and `IMAGE_TAG`. `make run` builds and grants the binary `NET_ADMIN` using sudo. `make clean` removes the default `bin/` directory.

The Dockerfile uses BuildKit cache mounts and cross-compiles for the target platform, including ARM v6/v7 variants. Its build arguments are `APP_VERSION`, `VCS_REF`, and `DATE_CREATED`; release automation supplies all three. Example:

```bash
docker build \
  --build-arg APP_VERSION=v1.4.0 \
  --build-arg VCS_REF="$(git rev-parse HEAD)" \
  --build-arg DATE_CREATED="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -t ipset-exporter:local .
docker run --rm ipset-exporter:local -version
```

## Troubleshooting

- **Operation not permitted:** check `CAP_NET_ADMIN`, binary file capabilities, and the network namespace. Rootless containers generally cannot monitor host ipsets.
- **Missing sets:** check `ipset.names` and run `ipset list -name` with the required permissions in the same namespace. Missing names are omitted, not emitted as zero.
- **Connection refused:** check the bind address, port, and logs. Listener failures terminate the exporter with a nonzero status.
- **Remote scrape fails:** the example now binds to loopback by default. Set `app.host` to an appropriate reachable address and check your firewall.

## Contributing

Open an issue or pull request on [GitHub](https://github.com/hatamiarash7/ipset-exporter). Include reproduction steps and run `make check test-race` before submitting changes. Use an isolated network namespace when testing real sets.

## Support

[Donate with Bitcoin](https://donatebadges.ir/donate/Bitcoin/bc1qmmh6vt366yzjt3grjxjjqynrrxs3frun8gnxrz) · [Donate with Ethereum](https://donatebadges.ir/donate/Ethereum/0x0831bD72Ea8904B38Be9D6185Da2f930d6078094) · [PayPing](https://payping.ir/@hatamiarash7)

## License

[MIT](LICENSE) © Arash Hatami.
