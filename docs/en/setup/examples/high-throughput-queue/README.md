# High throughput queue

High throughput queues can send messages to multiple channels using by `round-robin` policy for load-balance, and consume messages from each queue in parallel.
This prevents the single-threaded sender from performing network IO, blocking the `recieve`/`fetch`/`consumer` to the queue.

## Config

Both existing two queues(`memory`, `mmap`) support this future, add the `partition_count` to the queue config, default is `1`.

```yaml
queue:
    plugin_name: "memory-queue"
    # The maximum buffer event size.
    event_buffer_size: ${SATELLITE_QUEUE_EVENT_BUFFER_SIZE:5000}
    # The partition count of queue.
    partition_count: ${SATELLITE_QUEUE_PARTITION_COUNT:2}
```

Following the config, we create a partitioned queue, have two sub-queue, each sub-queue has 5000 buffer.
