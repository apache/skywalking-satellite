# Project Structure
- configs: Satellite configs.
- internal: Core, Api, and common utils.
- internal/pkg: Sharing with Core and Plugins, such as api and utils.
- internal/satellite: The core of Satellite.
- plugins: Contains all plugins.
- plugins/{type}: Contains the plugins of this {type}. Satellite has 6 plugin types, which are collector, queue, parser, filter, client, and forward.
- plugins/{type}/{plugin-name}: Contains the specific plugin, and {plugin-name}-{type} would be registered as the plugin unique name in the registry. 


```
.
├── configs
│   └── config.yaml
├── internal
│   ├── pkg
│   │   ├── api
│   │   │   ├── client.go
│   │   │   ├── collector.go
│   │   │   ├── event.go
│   │   │   ├── filter.go
│   │   │   ├── forwarder.go
│   │   │   ├── parser.go
│   │   │   ├── plugin.go
│   │   │   └── queue.go
│   │   └── ...
│   └── satellite
│       ├── registry
│       │   └── registry.go
│       └── ...
└── plugins
    ├── client
    │   ├── grpc
    │   │   └── README.md
    │   └── kakka
    │       └── README.md
    ├── collector
    │   └── log-grpc
    │       └── README.md
    ├── filter
    │   └── sampling
    │       └── README.md
    ├── forwarder
    │   └── segment
    │       └── README.md
    ├── parser
    │   └── gork
    │       └── README.md
    └── queue
        └── mmap
            └── README.md
```