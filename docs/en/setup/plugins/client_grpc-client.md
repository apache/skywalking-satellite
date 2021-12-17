# Client/grpc-client
## Description
The gRPC client is a sharing plugin to keep connection with the gRPC server and delivery the data to it.
## DefaultConfig
```yaml
# The gRPC client finder type
finder_type: "static"

# The gRPC server address (default localhost:11800), multiple addresses are split by ",".
server_addr: localhost:11800

# The gRPC kubernetes server address finder
kubernetes_config:
  # The kind of resource
  kind: pod
  # The resource namespaces
  namespaces:
    - default
  # How to get the address exported port
  extra_port:
    # Resource target port
    port: 11800

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

# How frequently to check the connection(second)
check_period: 5
```
## Configuration
|Name|Type|Description|
|----|----|-----------|
| finder_type | string | The gRPC server address finder type, support "static" and "kubernetes" |
| server_addr | string | The gRPC server address, only works for "static" address finder |
| kubernetes_config | *resolvers.KubernetesConfig | The kubernetes config to lookup addresses, only works for "kubernetes" address finder |
| kubernetes_config.api_server | string | The kubernetes API server address, If not define means using in kubernetes mode to connect |
| kubernetes_config.basic_auth | *resolvers.BasicAuth | The HTTP basic authentication credentials for the targets. |
| kubernetes_config.basic_auth.username | string |  |
| kubernetes_config.basic_auth.password | resolvers.Secret |  |
| kubernetes_config.basic_auth.password_file | string |  |
| kubernetes_config.bearer_token | resolvers.Secret | The bearer token for the targets. |
| kubernetes_config.bearer_token_file | string | The bearer token file for the targets. |
| kubernetes_config.proxy_url | string | HTTP proxy server to use to connect to the targets. |
| kubernetes_config.tls_config | resolvers.TLSConfig | TLSConfig to use to connect to the targets. |
| kubernetes_config.namespaces | []string | Support to lookup namespaces |
| kubernetes_config.kind | string | The kind of api |
| kubernetes_config.selector | resolvers.Selector | The kind selector |
| kubernetes_config.extra_port | resolvers.ExtraPort | How to get the address exported port |
| enable_TLS | bool | Enable TLS connect to server |
| client_pem_path | string | The file path of client.pem. The config only works when opening the TLS switch. |
| client_key_path | string | The file path of client.key. The config only works when opening the TLS switch. |
| ca_pem_path | string | The file path oca.pem. The config only works when opening the TLS switch. |
| insecure_skip_verify | bool | Controls whether a client verifies the server's certificate chain and host name. |
| authentication | string | The auth value when send request |
| check_period | int | How frequently to check the connection(second) |

