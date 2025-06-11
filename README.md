# Prometheus Exporter for ipset

[![Golang][golang_badge]][golang_link]
[![Release][release_badge]][release_link]
[![License][badge_license]][link_license]
[![Image size][badge_size_latest]][link_docker_hub]

It's a simple [ipset](https://linux.die.net/man/8/ipset) exporter that generate [Prometheus](https://prometheus.io/) metrics from ipset lists. Every metric has a label `set` that shows the ipset list name with the number of elements in the list as the value.

```text
ipset_entries_count{set="my-set-1",type="hash:ip"} 41
ipset_entries_count{set="my-set-2",type="hash:net"} 23
```

Additionally, the exporter exposes `ipset_update_errors_total`, a counter that tracks the number of errors encountered when trying to fetch ipset data.

## How-to 🚀

This exporter needs `NET_ADMIN` to fetch ipset data using netlink. You can build and run this exporter as a single binary:

```bash
make
sudo setcap cap_net_admin+ep ./bin/ipset-exporter
cp config.yml.example
./bin/ipset-exporter
```

Or you can use Docker:

```bash
docker run -d \
    --name ipset-exporter \
    --net host \
    -v /path/to/config.yml:/app/configs/config.yml \
    -e CONFIG_FILE=/app/configs/config.yml \
    --cap-add=NET_ADMIN \
    hatamiarash7/ipset-exporter:latest
```

> [!NOTE]
> The Docker container needs to run with `NET_ADMIN` capability to fetch ipset data.  
> Also, You should run it with `--net host` to access the host network namespace.

## Configuration 🛠

You can configure the exporter using a YAML file. The example configuration is:

```yaml
app:
    host: 127.0.0.1
    port: 4613
    log_level: info

ipset:
    update_interval: 15
    names:
        - my-set-1
        - my-set-2
```

-   You can choose any ipset name that you want to monitor or use the `all` keyword to monitor all ipset lists.
-   There is a `update_interval` field that set the interval in seconds to update the metrics.

---

## Support 💛

[![Donate with Bitcoin](https://img.shields.io/badge/Bitcoin-bc1qmmh6vt366yzjt3grjxjjqynrrxs3frun8gnxrz-orange)](https://donatebadges.ir/donate/Bitcoin/bc1qmmh6vt366yzjt3grjxjjqynrrxs3frun8gnxrz) [![Donate with Ethereum](https://img.shields.io/badge/Ethereum-0x0831bD72Ea8904B38Be9D6185Da2f930d6078094-blueviolet)](https://donatebadges.ir/donate/Ethereum/0x0831bD72Ea8904B38Be9D6185Da2f930d6078094)

<div><a href="https://payping.ir/@hatamiarash7"><img src="https://cdn.payping.ir/statics/Payping-logo/Trust/blue.svg" height="128" width="128"></a></div>

## Contributing 🤝

Don't be shy and reach out to us if you want to contribute 😉

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

## Issues

Each project may have many problems. Contributing to the better development of this project by reporting them. 👍

[golang_badge]: https://img.shields.io/badge/Made_With-Golang-blue
[golang_link]: https://go.dev/
[release_badge]: https://github.com/hatamiarash7/ipset-exporter/actions/workflows/release.yml/badge.svg
[release_link]: https://github.com/hatamiarash7/ipset-exporter/actions/workflows/release.yaml
[link_license]: https://github.com/hatamiarash7/ipset-exporter/blob/master/LICENSE
[badge_license]: https://img.shields.io/github/license/hatamiarash7/ipset-exporter.svg?longCache=true
[badge_size_latest]: https://img.shields.io/docker/image-size/hatamiarash7/ipset-exporter/latest?maxAge=30
[link_docker_hub]: https://hub.docker.com/r/hatamiarash7/ipset-exporter/
