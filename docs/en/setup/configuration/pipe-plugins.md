# Pipe Plugins
The pipe plugin configurations contain a series of pipe configuration. Each pipe configuration has 5 parts, which are common_config, gatherer, processor and the sender.

## common_config
|  Config   | Description  |
|  ----  | ----  |
| pipe_name  | The unique collect space name. |

## Gatherer
The gatherer has 2 roles, which are the receiver and fetcher.

### Receiver Role
|  Config   | Description  |
|  ----  | ----  |
| server_name  | The server name in the sharing pipe, which would be used in the receiver plugin.|
| receiver  | The receiver configuration. Please read [the doc](../plugins/plugin-list.md) to find all receiver plugins.|
| queue  | The queue buffers the input telemetry data. Please read [the doc](../plugins/plugin-list.md) to find all queue plugins.|



### Fetcher Role
|  Config   | Description  |
|  ----  | ----  |
| fetch_interval  | The time interval between two fetch operations. The time unit is millisecond.|
| fetcher  | The fetcher configuration. Please read [the doc](../plugins/plugin-list.md) to find all fetcher plugins.|
| queue  | The queue buffers the input telemetry data. Please read [the doc](../plugins/plugin-list.md) to find all queue plugins.|

## processor
The filter configuration. Please read [the doc](../plugins/plugin-list.md) to find all filter plugins.
## sender
|  Config   | Description  |
|  ----  | ----  |
| flush_time  | The time interval between two flush operations. And the time unit is millisecond.|
| max_buffer_size  | The maximum buffer elements.|
| min_flush_events  | The minimum flush elements.|
| client_name  | The client name used in the forwarders of the sharing pipe.|
| forwarders  |The forwarder plugin list. Please read [the doc](../plugins/plugin-list.md) to find all forwarders plugins.|
| fallbacker  |The fallbacker plugin. Please read [the doc](../plugins/plugin-list.md) to find all fallbacker plugins.|


## Example
```yaml
pipes:
  - common_config:
      pipe_name: pipe1
    gatherer:
      server_name: "grpc-server"
      receiver:
        plugin_name: "grpc-native-log-receiver"
      queue:
        plugin_name: "mmap-queue"
        segment_size: ${SATELLITE_MMAP_QUEUE_SIZE:524288}
        max_in_mem_segments: ${SATELLITE_MMAP_QUEUE_MAX_IN_MEM_SEGMENTS:6}
        queue_dir: "pipe1-log-grpc-receiver-queue"
    processor:
      filters:
    sender:
      fallbacker:
        plugin_name: none-fallbacker
      flush_time: ${SATELLITE_PIPE1_SENDER_FLUSH_TIME:1000}
      max_buffer_size: ${SATELLITE_PIPE1_SENDER_MAX_BUFFER_SIZE:200}
      min_flush_events: ${SATELLITE_PIPE1_SENDER_MIN_FLUSH_EVENTS:100}
      client_name: kafka-client
      forwarders:
        - plugin_name: native-log-kafka-forwarder
          topic: ${SATELLITE_NATIVELOG-TOPIC:log-topic}
```