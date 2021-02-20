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
```
