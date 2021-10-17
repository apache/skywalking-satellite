# Client/kafka-client
## Description
The Kafka client is a sharing plugin to keep connection with the Kafka brokers and delivery the data to it.
## DefaultConfig
```yaml
# The Kafka broker addresses (default localhost:9092). Multiple values are separated by commas.
brokers: localhost:9092

# The Kafka version should follow this pattern, which is major_minor_veryMinor_patch (default 1.0.0.0).
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
## Configuration
|Name|Type|Description|
|----|----|-----------|
| brokers | string | The Kafka broker addresses (default `localhost:9092`). |
| version | string | The version should follow this pattern, which is `major.minor.veryMinor.patch`. |
| enable_TLS | bool | The TLS switch (default false). |
| client_pem_path | string | The file path of client.pem. The config only works when opening the TLS switch. |
| client_key_path | string | The file path of client.key. The config only works when opening the TLS switch. |
| ca_pem_path | string | The file path oca.pem. The config only works when opening the TLS switch. |
| required_acks | int16 | 0 means NoResponse, 1 means WaitForLocal and -1 means WaitForAll (default 1). |
| producer_max_retry | int | The producer max retry times (default 3). |
| meta_max_retry | int | The meta max retry times (default 3). |
| retry_backoff | int | How long to wait for the cluster to settle between retries (default 100ms). |
| max_message_bytes | int | The max message bytes. |
| idempotent_writes | bool | Ensure that exactly one copy of each message is written when is true. |
| client_id | string | A user-provided string sent with every request to the brokers. |
| compression_codec | int | Represents the various compression codecs recognized by Kafka in messages. |
| refresh_period | int | How frequently to refresh the cluster metadata. |
| insecure_skip_verify | bool | Controls whether a client verifies the server's certificate chain and host name. |

