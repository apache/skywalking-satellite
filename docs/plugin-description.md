# api.Client
## kafka-client
### description
this is a sharing client to delivery the data to Kafka.
### defaultConfig
```yaml
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
refresh_period: 10

# InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
insecure_skip_verify: true
```
# api.Fallbacker
## none-fallbacker
### description
this is a nothing to do fallbacker.
### defaultConfig
```yaml```
# api.Fallbacker
## timer-fallbacker
### description
this is a timer fallback trigger when forward fails.
### defaultConfig
```yaml
max_times: 3
latency_factor: 2000
```
# api.Forwarder
## nativelog-kafka-forwarder
### description
this is a synchronization Kafka log forwarder.
### defaultConfig
```yaml
# The remote topic. 
topic: "log-topic"
```
# api.Queue
## memory-queue
### description
this is a memory queue to buffer the input event.
### defaultConfig
```yaml
# The maximum buffer event size.
event_buffer_size: 5000
# The discard strategy when facing the full condition.
# There are 2 strategies, which are LOST_THE_OLDEST_ONE and LOST_THE_NEW_ONE. 
discard_strategy: LOST_THE_OLDEST_ONE
```
# api.Queue
## mmap-queue
### description
this is a memory mapped queue to provide the persistent storage.
### defaultConfig
```yaml
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
# api.Receiver
## grpc-nativelog-receiver
### description
This is a receiver for SkyWalking native logging format, which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto.
### defaultConfig
```yaml```
# api.Receiver
## http-log-receiver
### description
This is a receiver for SkyWalking http logging format, which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto.
### defaultConfig
```yaml
# The native log request URI.
uri: "/logging"
# The request timeout seconds.
timeout: 5
```
# api.Server
## grpc-server
### description
this is a grpc server
### defaultConfig
```yaml
# The address of grpc server. Default value is :8000
address: :8000
# The network of grpc. Default value is :tcp
network: tcp
# The max size of receiving log. Default value is 2M. The unit is Byte.
max_recv_msg_size: 2097152
# The max concurrent stream channels.
max_concurrent_streams: 32
# The TLS cert file path.
tls_cert_file: 
# The TLS key file path.
tls_key_file: 
```
# api.Server
## http-server
### description
this is a http server.
### defaultConfig
```yaml
# The http server address.
address: ":8080"
```
# api.Server
## prometheus-server
### description
this is a prometheus server to export the metrics in Satellite.
### defaultConfig
```yaml
# The prometheus server address.
address: ":9299"
# The prometheus server metrics endpoint.
endpoint: "/metrics"
```
