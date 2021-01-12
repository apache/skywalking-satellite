# api.Client
## kafka-client
### description
```this is a sharing client to delivery the data to Kafka.```
### defaultConfig
```
# The Kafka broker addresses (default localhost:9092). Multiple values are separated by commas.
brokers: localhost:9092

# The Kakfa version should follow this pattern, which is major_minor_veryMinor_patch (default 1.0.0.0).
version: 1.0.0.0

# The TLS switch (default false).
enable_TLS: false

# The file path of client.pem. The config only works when opening the TLS switch.
client_pem_path: ""

# The file path of client.key. The config only works when opening the TLS switch.
client_key_path: ""

# The file path oca.pem. The config only works when opening the TLS switch.
ca_pem_path: ""

# 0 means NoResponse, 1 means WaitForLocal and -1 means WaitForAll (default 1).
required_acks: 1

# The producer max retry times (default 3).
producer_max_retry: 3

# The meta max retry times (default 3).
meta_max_retry: 3

# How long to wait for the cluster to settle between retries (default 100ms). Time unit is ms.
retry_backoff: 100

# The max message bytes.
max_message_bytes: 1000000

# If enabled, the producer will ensure that exactly one copy of each message is written (default false).
idempotent_writes: false

# A user-provided string sent with every request to the brokers for logging, debugging, and auditing purposes (default Satellite).
client_id: Satellite

# Compression codec represents the various compression codecs recognized by Kafka in messages. 0 : None, 1 : Gzip, 2 : Snappy, 3 : LZ4, 4 : ZSTD
compression_codec: 0

# How frequently to refresh the cluster metadata in the background. Defaults to 10 minutes. The unit is minute.
# refresh_period: 10

# InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
insecure_skip_verify: true
```
# api.Fallbacker
## timer-fallbacker
### description
```this is a timer fallback trigger when forward fails.```
### defaultConfig
```
max_times: 3
latency_factor: 2000
```
# api.Forwarder
## log-kafka-forwarder
### description
```this is a synchronization Kafka log forwarder.```
### defaultConfig
```
# The remote topic. 
topic: "log-topic"
```
# api.Queue
## mmap-queue
### description
```this is a memory mapped queue to provide the persistent storage.```
### defaultConfig
```
# The size of each segment. Default value is 128K. The unit is Byte.
segment_size: 131072
# The max num of segments in memory. Default value is 10.
max_in_mem_segments: 10
# The capacity of Queue = segment_size * queue_capacity_segments.
queue_capacity_segments: 4000
# The period flush time. The unit is ms. Default value is 1 second.
flush_period: 1000
# The max number in one flush time.  Default value is 10000.
flush_ceiling_num: 10000
# Contains all files in the queue.
queue_dir: satellite-mmap-queue
# The max size of the input event. Default value is 20k.
max_event_size: 20480
```
# api.Server
## prometheus-server
### description
```this is a prometheus server to export the metrics in Satellite.```
### defaultConfig
```
# The prometheus server address.
address: ":9299"
# The prometheus server metrics endpoint.
endpoint: "/metrics"
```
