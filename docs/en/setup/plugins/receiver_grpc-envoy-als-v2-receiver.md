# Receiver/grpc-envoy-als-v2-receiver
## Description
This is a receiver for Envoy ALS format, which is defined at https://github.com/envoyproxy/envoy/blob/v1.17.4/api/envoy/service/accesslog/v2/als.proto.
## Support Forwarders
 - [envoy-als-v2-grpc-forwarder](forwarder_envoy-als-v2-grpc-forwarder.md)
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

