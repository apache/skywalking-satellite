# Forwarder/otlp-metrics-v1-grpc-forwarder
## Description
This is a synchronization grpc forwarder with the OpenTelemetry metrics v1 protocol.
## DefaultConfig
```yaml
# The LRU policy cache size for hosting routine rules of service instance.
routing_rule_lru_cache_size: 5000
# The TTL of the LRU cache size for hosting routine rules of service instance.
routing_rule_lru_cache_ttl: 180
# The label key of the routing data, multiple keys are split by ","
routing_label_keys: net.host.name,host.name,job,service.name
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| routing_label_keys | string | The label key of the routing data, multiple keys are split by "," |
| routing_rule_lru_cache_size | int | The LRU policy cache size for hosting routine rules of service instance. |
| routing_rule_lru_cache_ttl | int | The TTL of the LRU cache size for hosting routine rules of service instance. |

