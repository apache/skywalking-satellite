# Project Structure
- configs: Satellite configs.
- internal: Core, Api, and common utils.
- internal/pkg: Sharing with Core and Plugins, such as api and utils.
- internal/satellite: The core of Satellite.
- plugins: Contains all plugins.
- plugins/{type}: Contains the plugins of this {type}. Satellite has 6 plugin types, which are collector, queue, parser, filter, client, and forward.
- plugins/api: Contains the plugin definition.
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
├── plugins
│   ├── client
│   │   ├── api
│   │   │   └── client.go
│   │   ├── grpc
│   │   └── kakka
│   ├── collector
│   │   ├── api
│   │   │   └── collector.go
│   │   ├── example
│   │   └── log-grpc
│   │       └── README.md
│   ├── fallbacker
│   │   ├── api
│   │   │   └── fallbacker.go
│   ├── filter
│   │   ├── api
│   │   │   └── filter.go
│   ├── forwarder
│   │   ├── api
│   │   │   └── forwarder.go
│   ├── parser
│   │   ├── api
│   │   │   └── parser.go
│   │   └── gork
│   │       └── README.md
│   └── queue
│       ├── api
│       │   └── queue.go
│       └── mmap
│           └── README.md
```
