# Forwarder/nativemeter-grpc-forwarder
## Description
This is a synchronization meter grpc forwarder with the SkyWalking meter protocol.
## DefaultConfig
```yaml
# The upstream LRU Cache max size
upstream_lru_cache_size: 5000
# The upstream LRU Cache time(second) on each service instance
upstream_lru_cache_ttl: 180
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| upstream_lru_cache_size | int | The upstream LRU Cache max size |
| upstream_lru_cache_ttl | int | The upstream LRU Cache time(second) on each service instance |

