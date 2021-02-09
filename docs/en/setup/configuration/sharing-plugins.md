# Sharing Plugins
Sharing plugin configurations has three 3 parts, which are common_config, clients and servers.


## Common Configuration
|  Config   |Default| Description  |
|  ----  | ----  | ----  |
| pipe_name  | sharing| The group name of sharing plugins |

## Clients
Clients have a series of client plugins, which would be sharing with the plugins of the other pipes. Please read [the doc](../plugins/plugin-list.md) to find all client plugin configurations.
## Servers
Servers have a series of server plugins, which would be sharing with the plugins of the other pipes. Please read [the doc](../plugins/plugin-list.md) to find all server plugin configurations.

## Example
```yaml
# The sharing plugins referenced by the specific plugins in the different pipes.
sharing:
  common_config:
    pipe_name: sharing
  clients:
    - plugin_name: "kafka-client"
      brokers: ${SATELLITE_KAFKA_CLIENT_BROKERS:127.0.0.1:9092}
      version: ${SATELLITE_KAFKA_VERSION:"2.1.1"}
  servers:
    - plugin_name: "grpc-server"
    - plugin_name: "prometheus-server"
      address: ${SATELLITE_PROMETHEUS_ADDRESS:":8090"}
```