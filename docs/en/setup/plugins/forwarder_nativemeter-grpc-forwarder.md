# Forwarder/nativemeter-grpc-forwarder
## Description
This is a synchronization meter grpc forwarder with the SkyWalking meter protocol.
## DefaultConfig
```yaml
# The LRU policy cache size for hosting routine rules of service instance.
routing_rule_lru_cache_size: 5000
# The TTL of the LRU cache size for hosting routine rules of service instance.
routing_rule_lru_cache_ttl: 180
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| routing_rule_lru_cache_size | int | The LRU policy cache size for hosting routine rules of service instance. |
| routing_rule_lru_cache_ttl | int | The TTL of the LRU cache size for hosting routine rules of service instance. |

