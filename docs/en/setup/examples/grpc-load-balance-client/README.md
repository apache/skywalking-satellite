# GRPC Load Balance Client

GRPC client support connect to multiple server address, and use `round-robin` policy for load-balance server before send each request.

## Server Discovery

Support two ways to locate the server list, using `finder_type` to change the type to find:
1. `static`: Define the server address list.
2. `kubernetes`: Define kubernetes pod/service/endpoint, it could be found addresses and dynamic update automatically.

### Static server list

You could see there define two server address and split by ",".

```yaml
sharing:
  clients:
    - plugin_name: "grpc-client"
      # The gRPC server address finder type
      finder_type: ${SATELLITE_GRPC_CLIENT_FINDER:static}
      # The gRPC server address (default localhost:11800).
      server_addr: ${SATELLITE_GRPC_CLIENT:127.0.0.1:11800,127.0.0.2:11800}
      # The TLS switch
      enable_TLS: ${SATELLITE_GRPC_ENABLE_TLS:false}
      # The file path of client.pem. The config only works when opening the TLS switch.
      client_pem_path: ${SATELLITE_GRPC_CLIENT_PEM_PATH:"client.pem"}
      # The file path of client.key. The config only works when opening the TLS switch.
      client_key_path: ${SATELLITE_GRPC_CLIENT_KEY_PATH:"client.key"}
      # InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
      insecure_skip_verify: ${SATELLITE_GRPC_INSECURE_SKIP_VERIFY:false}
      # The file path oca.pem. The config only works when opening the TLS switch.
      ca_pem_path: ${SATELLITE_grpc_CA_PEM_PATH:"ca.pem"}
      # How frequently to check the connection(second)
      check_period: ${SATELLITE_GRPC_CHECK_PERIOD:5}
      # The auth value when send request
      authentication: ${SATELLITE_GRPC_AUTHENTICATION:""}
      address: ${SATELLITE_GRPC_ADDRESS:":11800"}
      # The TLS cert file path.
      tls_cert_file: ${SATELLITE_GRPC_TLS_KEY_FILE:""}
      # The TLS key file path.
      tls_key_file: ${SATELLITE_GRPC_TLS_KEY_FILE:""}
```

### Kubernetes selector

Using `kubernetes_config` to define the address's finder.

```yaml
sharing:
  clients:
    - plugin_name: "grpc-client"
      # The gRPC server address finder type
      finder_type: ${SATELLITE_GRPC_CLIENT_FINDER:kubernetes}
      # The kubernetes config to lookup addresses
      kubernetes_config:
        # The kubernetes API server address, If not define means using in kubernetes mode to connect
        api_server: http://localhost:8001/
        # The kind of api
        kind: endpoints
        # Support to lookup namespaces
        namespaces:
          - default
        # The kind selector
        selector:
          label: app=productpage
        # How to get the address exported port
        extra_port:
          port: 9080
      # The TLS switch
      enable_TLS: ${SATELLITE_GRPC_ENABLE_TLS:false}
      # The file path of client.pem. The config only works when opening the TLS switch.
      client_pem_path: ${SATELLITE_GRPC_CLIENT_PEM_PATH:"client.pem"}
      # The file path of client.key. The config only works when opening the TLS switch.
      client_key_path: ${SATELLITE_GRPC_CLIENT_KEY_PATH:"client.key"}
      # InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
      insecure_skip_verify: ${SATELLITE_GRPC_INSECURE_SKIP_VERIFY:false}
      # The file path oca.pem. The config only works when opening the TLS switch.
      ca_pem_path: ${SATELLITE_grpc_CA_PEM_PATH:"ca.pem"}
      # How frequently to check the connection(second)
      check_period: ${SATELLITE_GRPC_CHECK_PERIOD:5}
      # The auth value when send request
      authentication: ${SATELLITE_GRPC_AUTHENTICATION:""}
      address: ${SATELLITE_GRPC_ADDRESS:":11800"}
      # The TLS cert file path.
      tls_cert_file: ${SATELLITE_GRPC_TLS_KEY_FILE:""}
      # The TLS key file path.
      tls_key_file: ${SATELLITE_GRPC_TLS_KEY_FILE:""}
```
