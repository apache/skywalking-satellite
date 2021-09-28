# Server/prometheus-server
## Description
This is a prometheus server to export the metrics in Satellite.
## DefaultConfig
```yaml
# The prometheus server address.
address: ":1234"
# The prometheus server metrics endpoint.
endpoint: "/metrics"
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| address | string | The prometheus server address. |
| endpoint | string | The prometheus server metrics endpoint. |

