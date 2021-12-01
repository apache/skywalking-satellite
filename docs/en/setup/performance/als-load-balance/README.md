# ALS Load Balance

Using satellite as a load balancer in envoy and OAP can effectively prevent the problem of unbalanced messages received by OAP.

In this case, we mainly use memory queues for intermediate data storage. 

Deference Envoy Count, OAP performance could impact the Satellite transmit performance.

|Envoy Instance|Concurrent User|ALS OPS|Satellite CPU|Satellite Memory|
|--------------|---------------|-------|-------------|----------------|
|150|100|~50K|1.2C|0.5-1.0G|
|150|300|~80K|1.8C|1.0-1.5G|
|300|100|~50K|1.4C|0.8-1.2G|
|300|300|~100K|2.2C|1.3-2.0G|
|800|100|~50K|1.5C|0.9-1.5G|
|800|300|~100K|2.6C|1.7-2.7G|
|1500|100|~50K|1.7C|1.4-2.4G|
|1500|300|~100K|2.7C|2.3-3.0G|
|2300|150|~50K|1.8C|1.9-3.1G|
|2300|300|~90K|2.5C|2.3-4.0G|
|2300|500|~110K|3.2C|2.8-4.7G|

## Detail

### Environment

Using GKE Environment, helm to build cluster.

|Module|Version|Replicate Count|CPU Limit|Memory Limit|Description|
|------|-------|---------------|---------|------------|-----------|
|OAP|8.9.0|6|12C|32Gi|Using ElasticSearch as Storage|
|Satellite|0.4.0|1|8C|16Gi||
|ElasticSearch|7.5.1|3|8|16Gi||

### Setting

800 Envoy, 100K QPS ALS.

|Module|Environment Config|Use Value|Default Value|Description|Recommend Value|
|------|------------------|---------|-------------|-----------|--------------|
|Satellite|SATELLITE_QUEUE_PARTITION|50|4|Support several goroutines concurrently to consume the queue|Satellite CPU number * 4-6, It could help improve throughput, but the default value also could handle `800` Envoy Instance and `100K` QPS ALS message. |
|Satellite|SATELLITE_QUEUE_EVENT_BUFFER_SIZE|3000|1000|The size of the queue in each concurrency|This is related to the number of Envoys. If the number of Envoys is large, it is recommended to increase the value.|
|Satellite|SATELLITE_ENVOY_ALS_V3_PIPE_RECEIVER_FLUSH_TIME|3000|1000|When the Satellite receives the message, how long(millisecond) will the ALS message be merged into an Event.|If a certain time delay is accepted, the value can be adjusted larger, which can effectively reduce CPU usage and make the Satellite more stable|
|Satellite|SATELLITE_ENVOY_ALS_V3_PIPE_SENDER_FLUSH_TIME|3000|1000|How long(millisecond) is the memory queue data for each Goroutine to be summarized and sent to OAP|This depends on the amount of data in your queue, you can keep it consistent with `SATELLITE_ENVOY_ALS_V3_PIPE_RECEIVER_FLUSH_TIME`|
|OAP|SW_CORE_GRPC_MAX_CONCURRENT_CALL|50|4|A link between Satellite and OAP, how many requests parallelism is supported|Same with `SATELLITE_QUEUE_PARTITION` in Satellite|
