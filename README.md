# Prometheus Exporter for ipset

It's a simple [ipset](https://linux.die.net/man/8/ipset) exporter that generate [Prometheus](https://prometheus.io/) metrics from ipset lists. Every metric has a label `set` that shows the ipset list name with the number of elements in the list as the value.

```text
ipset_count{set="my-set-1"} 41
ipset_count{set="my-set-2"} 23
```

## How-to

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
    -p 4613:4613 \
    -v /path/to/config.yml:/app/configs/config.yml \
    -e CONFIG_FILE=/app/configs/config.yml \
    --cap-add=NET_ADMIN \
    hatamiarash7/ipset-exporter:latest
```

## Configuration

You can configure the exporter using a YAML file. The example configuration is:

```yaml
app:
    host: 127.0.0.1
    port: 4613
    log_level: info

ipset:
    names:
        - my-set-1
        - my-set-2
```

You can choose any ipset name that you want to monitor or use the `all` keyword to monitor all ipset lists.

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
