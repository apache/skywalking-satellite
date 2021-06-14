# Fetcher/prometheus-fetcher
## Description
This is a Prometheus fetcher for SkyWalking meter format, which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/language-agent/Meter.proto.
## DefaultConfig
```yaml
## Prometheus scrape configure
scrape_configs:
  - job_name: 'prometheus'
    metrics_path: '/metrics'
    scrape_interval: 10s
    static_configs:
      - targets: ['127.0.0.1:2020']
```
