# Fetcher/prometheus-metrics-fetcher

## Description
This is a fetcher for Skywalking prometheus metrics format, which will translate Prometheus metrics to Skywalking meter system.

## Configuration

Support using static configuration or Kubernetes service discovery configuration to scrape Prometheus.

```yaml
gatherer:
  fetcher:
    plugin_name: "prometheus-metrics-fetcher"
    scrape_configs:
      # Static configuration
      - job_name: 'prometheus-static'
        metrics_path: '/metrics'
        scrape_interval: 10s
        static_configs:
          - targets:
              - "127.0.0.1:9100"
      # Kubernetes service discovery configuration
      - job_name: 'prometheus-k8s'
        metrics_path: '/metrics'
        scrape_interval: 10s
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        kubernetes_sd_configs:
          - role: pod
            selectors:
              - role: pod
                label: "app=prometheus"
```
