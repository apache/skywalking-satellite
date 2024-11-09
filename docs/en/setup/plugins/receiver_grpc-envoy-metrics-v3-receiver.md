# Receiver/grpc-envoy-metrics-v3-receiver
## Description
This is a receiver for Envoy Metrics format, which is defined at https://github.com/envoyproxy/envoy/blob/5f7d6efb5786ee3de31b1fb37c78fa281718b704/api/envoy/service/metrics/v3/metrics_service.proto.
## Support Forwarders
 - [envoy-metrics-v3-grpc-forwarder](forwarder_envoy-metrics-v3-grpc-forwarder.md)
## DefaultConfig
```yaml
# The time interval between two flush operations. And the time unit is millisecond.
flush_time: 1000
# The max cache count when receive the message
limit_count: 500
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| flush_time | int | The time interval between two flush operations. And the time unit is millisecond. |
| limit_count | int | The max cache count when receive the message |

