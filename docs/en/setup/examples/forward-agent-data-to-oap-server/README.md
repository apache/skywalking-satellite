# Forward agent data to OAP server

Using Satellite to receive the SkyWalking protocols from agent, and transport data to the SkyWalking backend or another Satellite.

## Protocols 
Support these protocols transport via `gRPC`:
1. Tracing
2. Log
3. Management 
4. CDS 
5. Event 
6. JVM 
7. Profile

## Config

Here is [config file](satellite_config.yaml), set out as follows:

- Declare gRPC [server](../../plugins/server_grpc-server.md) and [client](../../plugins/client_grpc-client.md) to receive and transmit data.
- Declare each protocol gatherer and sender to transmit protocol via [pipeline](../../configuration/pipe-plugins.md).
- Expose [Self-Observability telemetry data](../../configuration/common.md#self-telemetry) to [Prometheus](../../plugins/server_prometheus-server.md).