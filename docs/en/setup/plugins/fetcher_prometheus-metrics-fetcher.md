# Fetcher/prometheus-metrics-fetcher
## Description
This is a fetcher for Skywalking prometheus metrics format, which will translate Prometheus metrics to Skywalking meter system.
## DefaultConfig
```yaml
scrape_configs:
 - job_name: 'prometheus'
   metrics_path: '/metrics'
   scrape_interval: 10s
   static_configs:
   - targets: ['127.0.0.1:9100']
```
