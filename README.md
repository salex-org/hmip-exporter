# HomematicIP Exporter

A Prometheus Metrics Exporter for the HomematicIP Cloud. This library is an unofficial implementation
based on the [HmIP Go Client](https://github.com/salex-org/hmip-go-client) and is not affiliated with
eQ-3 AG!

The current version is work in progress that does not cover all functions and has not been fully tested.
**Use this exporter at your own risk!**

## Build locally

Build and run locally on MacOS:

```shell
goreleaser release --snapshot --clean
docker run -d -p 9100:9100 ghcr.io/salex-org/hmip-exporter:latest-arm64
```

Call metrics endpoint locally:

```shell
curl http://localhost:9100/metrics
```