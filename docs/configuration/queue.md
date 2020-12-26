# queue configuration

|  Type   | Param  | DefaultValue| Meaning| 
|  ----  | ----  |----  | ----  |
| mmap-queue  | segment_size | 131072 | The size of each segment(Unit:Byte). The minimum value is the system memory page size.
| mmap-queue  | max_in_mem_segments | 10 | The max num of segments in memory. The minimum value is 4.
| mmap-queue  | queue_capacity_segments | 4000 | The capacity of Queue = segment_size * queue_capacity_segments.
| mmap-queue  | flush_period | 1000 | The period flush time. The unit is ms.
| mmap-queue  | flush_ceiling_num | 10000 | The max number in one flush time.
| mmap-queue  | queue_dir | satellite-mmap-queue |Contains all files in the queue.
| mmap-queue  | max_event_size | 20480 |The max size of the input event(Unit:Byte).
