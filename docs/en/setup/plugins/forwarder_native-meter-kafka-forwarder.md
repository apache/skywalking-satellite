# Forwarder/native-meter-kafka-forwarder
## Description
This is a synchronization Kafka forwarder with the SkyWalking native meter protocol.
## DefaultConfig
```yaml
# The remote topic. 
topic: "skywalking-meters"
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| routing_rule_lru_cache_size | int |  |
| routing_rule_lru_cache_ttl | int | The TTL of the LRU cache size for hosting routine rules of service instance. |
| topic | string | The forwarder topic. |

