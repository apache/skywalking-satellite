# Queue/mmap-queue
## Description
This is a memory mapped queue to provide the persistent storage for the input event. Please note that this plugin does not support Windows platform.
## DefaultConfig
```yaml
# The size of each segment. Default value is 256K. The unit is Byte.
segment_size: 262114
# The max num of segments in memory. Default value is 10.
max_in_mem_segments: 10
# The capacity of Queue = segment_size * queue_capacity_segments.
queue_capacity_segments: 2000
# The period flush time. The unit is ms. Default value is 1 second.
flush_period: 1000
# The max number in one flush time.  Default value is 10000.
flush_ceiling_num: 10000
# The max size of the input event. Default value is 20k.
max_event_size: 20480

# The partition count of queue.
partition_count: 1
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| segment_size | int | The size of each segment. The unit is byte. |
| max_in_mem_segments | int32 | The max num of segments in memory. |
| queue_capacity_segments | int | The capacity of Queue = segment_size * queue_capacity_segments. |
| flush_period | int | The period flush time. The unit is ms. |
| flush_ceiling_num | int | The max number in one flush time. |
| max_event_size | int | The max size of the input event. |
| partition_count | int | The total partition count. |

