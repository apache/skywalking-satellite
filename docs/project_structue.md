# Project Structure
- configs: Satellite configs.
- internal: Core, Api, and common utils.
- internal/pkg: Sharing with Core and Plugins, such as api and utils.
- internal/satellite: The core of Satellite.
- plugins: Contains all plugins.
- plugins/{type}: Contains the plugins of this {type}. Satellite has 9 plugin types.
- plugins/api: Contains the plugin definition and initlizer.
- plugins/{type}/{plugin-name}: Contains the specific plugin, and {plugin-name}-{type} would be registered as the plugin unique name in the registry. 
```
.
├── cmd
│   ├── command.go
│   └── main.go
├── configs
│   └── satellite_config.yaml
├── docs
│   ├── design
│   │   ├── module_design.md
│   │   ├── module_structure.md
│   │   └── plugin_structure.md
│   └── project_structue.md
├── internal
│   ├── container
│   ├── pkg
│   │   ├── event
│   │   ├── log
│   │   ├── plugin
│   │   │   ├── definition.go
│   │   │   ├── plugin_test.go
│   │   │   └── registry.go
│   └── satellite
│       ├── boot
│       │   └── boot.go
│       ├── config
│       ├── event
│       ├── module
│       │   ├── api
│       │   ├── buffer
│       │   ├── gatherer
│       │   ├── processor
│       │   └── sender
│       └── sharing
└── plugins
    ├── client
    │   ├── api
    │   │   ├── client.go
    │   │   └── client_repository.go
    ├── fallbacker
    │   ├── api
    │   │   ├── fallbacker.go
    │   │   └── fallbacker_repository.go
    ├── fetcher
    │   └── api
    │       ├── fetcher.go
    │       └── fetcher_repository.go
    ├── filter
    │   ├── api
    │   │   ├── filter.go
    │   │   └── filter_repository.go
    ├── forwarder
    │   ├── api
    │   │   ├── forwarder.go
    │   │   └── forwarder_repository.go
    ├── init.go
    ├── parser
    │   ├── api
    │   │   ├── parser.go
    │   │   └── parser_repository.go
    ├── queue
    │   ├── api
    │   │   ├── queue.go
    │   │   └── queue_repository.go
    ├── receiver
    │   ├── api
    │   │   ├── receiver.go
    │   │   └── receiver_repository.go
    └── server
        └── api
            ├── server.go
            └── server_repository.go
```
