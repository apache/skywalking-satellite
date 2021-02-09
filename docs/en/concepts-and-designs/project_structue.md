# Project Structure
- cmd: The starter of Satellite.
- configs: Satellite configs.
- internal: Core, Api, and common utils.
    - internal/pkg: Sharing with Core and Plugins, such as api and utils.
    - internal/satellite: The core of Satellite.
- plugins: Contains all plugins.
    - plugins/{type}: Contains the plugins of this {type}. Satellite has 9 plugin types.
    - plugins/{type}/api: Contains the plugin definition and initializer.
    - plugins/{type}/{plugin-name}: Contains the specific plugin.
    - init.go: Register the plugins to the plugin registry.
```
.
├── CHANGES.md
├── cmd
├── configs
├── docs
├── go.sum
├── internal
│   ├── pkg
│   └── satellite
├── plugins
│   ├── client
│   ├── fallbacker
│   ├── fetcher
│   ├── filter
│   ├── forwarder
│   ├── init.go
│   ├── parser
│   ├── queue
│   ├── receiver
│   └── server
```
