# Receiver/grpc-envoy-als-v3-receiver
## Description
This is a receiver for Envoy ALS format, which is defined at https://github.com/envoyproxy/envoy/blob/3791753e94edbac8a90c5485c68136886c40e719/api/envoy/config/accesslog/v3/accesslog.proto.
## Support Forwarders
 - [envoy-als-v3-grpc-forwarder](forwarder_envoy-als-v3-grpc-forwarder.md)
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

