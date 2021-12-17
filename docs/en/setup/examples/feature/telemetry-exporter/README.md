# Telemetry Exporter

Satellite supports two ways to export its own telemetry data, `prometheus` or `metrics-service`.

## Prometheus

Start HTTP port to export the satellite telemetry metrics. 

When the following configuration is completed, then the satellite telemetry metrics export to: `http://localhost${SATELLITE_TELEMETRY_PROMETHEUS_ADDRESS}${SATELLITE_TELEMETRY_PROMETHEUS_ENDPOINT}`,
and all the metrics contain the `cluster`, `service` and `instance` tag.

```xml
# The Satellite self telemetry configuration.
telemetry:
  # The space concept for the deployment, such as the namespace concept in the Kubernetes.
  cluster: ${SATELLITE_TELEMETRY_CLUSTER:satellite-cluster}
  # The group concept for the deployment, such as the service resource concept in the Kubernetes.
  service: ${SATELLITE_TELEMETRY_SERVICE:satellite-service}
  # The minimum running unit, such as the pod concept in the Kubernetes.
  instance: ${SATELLITE_TELEMETRY_SERVICE:satellite-instance}
  # Telemetry export type, support "prometheus", "metrics_service" or "none"
  export_type: ${SATELLITE_TELEMETRY_EXPORT_TYPE:prometheus}
  # Export telemetry data through Prometheus server, only works on "export_type=prometheus".
  prometheus:
    # The prometheus server address.
    address: ${SATELLITE_TELEMETRY_PROMETHEUS_ADDRESS::1234}
    # The prometheus server metrics endpoint.
    endpoint: ${SATELLITE_TELEMETRY_PROMETHEUS_ENDPOINT:/metrics}
```

## Metrics Service

Send the message to the gRPC service that supports SkyWalking's native Meter protocol with interval.

When the following configuration is completed, send the message to the specified `grpc-client` component at the specified time interval.
Among them, `service` and `instance` will correspond to the services and service instances in SkyWalking.

```xml

# The Satellite self telemetry configuration.
telemetry:
  # The space concept for the deployment, such as the namespace concept in the Kubernetes.
  cluster: ${SATELLITE_TELEMETRY_CLUSTER:satellite-cluster}
  # The group concept for the deployment, such as the service resource concept in the Kubernetes.
  service: ${SATELLITE_TELEMETRY_SERVICE:satellite-service}
  # The minimum running unit, such as the pod concept in the Kubernetes.
  instance: ${SATELLITE_TELEMETRY_SERVICE:satellite-instance}
  # Telemetry export type, support "prometheus", "metrics_service" or "none"
  export_type: ${SATELLITE_TELEMETRY_EXPORT_TYPE:metrics_service}
  # Export telemetry data through native meter format to OAP backend, only works on "export_type=metrics_service".
  metrics_service:
    # The grpc-client plugin name, using the SkyWalking native batch meter protocol
    client_name: ${SATELLITE_TELEMETRY_METRICS_SERVICE_CLIENT_NAME:grpc-client}
    # The interval second for sending metrics
    interval: ${SATELLITE_TELEMETRY_METRICS_SERVICE_INTERVAL:10}
```