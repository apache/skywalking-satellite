# Server/grpc-server
## Description
This is a sharing plugin, which would start a gRPC server.
## DefaultConfig
```yaml
# The address of grpc server. Default value is :11800
address: :11800
# The network of grpc. Default value is :tcp
network: tcp
# The max size of receiving log. Default value is 2M. The unit is Byte.
max_recv_msg_size: 2097152
# The max concurrent stream channels.
max_concurrent_streams: 32
# The TLS cert file path.
tls_cert_file: ""
# The TLS key file path.
tls_key_file: ""
# To Accept Connection Limiter when reach the resource
accept_limit:
  # The max CPU utilization limit
  cpu_utilization: 75
  # The max connection count
  connection_count: 4000
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| address | string | The address of grpc server. |
| network | string | The network of grpc. |
| max_recv_msg_size | int | The max size of the received log. |
| max_concurrent_streams | uint32 | The max concurrent stream channels. |
| tls_cert_file | string | The TLS cert file path. |
| tls_key_file | string | The TLS key file path. |
| accept_limit | grpc.AcceptConnectionConfig | To Accept Connection Limiter when reach the resource |

