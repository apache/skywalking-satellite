# Transmit Log to Kafka

Using Satellite to receive the SkyWalking log protocol from agent, and transport data to the Kafka Topic.

## Config

Here is [config file](satellite_config.yaml), set out as follows:

- Declare gRPC [server](../../plugins/server_grpc-server.md) and [kafka client](../../plugins/client_kafka-client.md) to receive and transmit data.
- Declare the SkyWalking Log protocol gatherer and sender to transmit protocol via [pipeline](../../configuration/pipe-plugins.md).
- Expose [Self-Observability telemetry data](../../configuration/common.md#self-telemetry) to [Prometheus](../../plugins/server_prometheus-server.md).