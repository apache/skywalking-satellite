# Queue/memory-queue
## Description
This is a memory queue to buffer the input event.
## DefaultConfig
```yaml
# The maximum buffer event size.
event_buffer_size: 5000

# The partition count of queue.
partition_count: 1
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| event_buffer_size | int | configThe maximum buffer event size. |
| partition_count | int | The total partition count. |

