# High throughput queue

High throughput queues can send messages to multiple channels by using `round-robin` policy for load-balance, and consume messages from each queue in parallel.
This prevents the single-threaded sender from performing network IO, blocking the `receiver`/`fetcher` to the queue.

## Config

Both existing two queues(`memory`, `mmap`) support this feature, add the `partition` to the queue config, default is `1`.

```yaml
queue:
    plugin_name: "memory-queue"
    # The maximum buffer event size.
    event_buffer_size: ${SATELLITE_QUEUE_EVENT_BUFFER_SIZE:5000}
    # The partition count of queue.
    partition: ${SATELLITE_QUEUE_PARTITION:2}
```

Following the config, we create a partitioned queue, have two sub-queue, each sub-queue has 5000 buffer.
