# Client/grpc-client
## Description
The gRPC client is a sharing plugin to keep connection with the gRPC server and delivery the data to it.
## DefaultConfig
```yaml
# The gRPC server address (default localhost:11800). 
server_addr: localhost:11800

# The TLS switch (default false).
enable_TLS: false

# The file path of client.pem. The config only works when opening the TLS switch.
client_pem_path: ""

# The file path of client.key. The config only works when opening the TLS switch.
client_key_path: ""

# The file path oca.pem. The config only works when opening the TLS switch.
ca_pem_path: ""

# InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
insecure_skip_verify: true

# The auth value when send request
authentication: ""

# How frequently to check the connection
check_period: 5
```
