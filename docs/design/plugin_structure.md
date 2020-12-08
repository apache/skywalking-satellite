# plugin structure
`Plugin is a common concept for Satellite. Not only does the extension mechanism depend on Plugin, but the core modules also depend on Plugin`

## Registration mechanism

The Plugin registration mechanism in Satellite is similar to the SPI registration mechanism of Java. 
Plugin registration mechanism supports to register an interface and its implementation, that means different interfaces have different registration spaces.
We can easily find the type of a specific plugin according to the interface and the plugin name and initialize it according to the type.

structure:
- code: map[reflect.Type]map[string]reflect.Value
- meaning: map[`interface type`]map[`plugin name`] `plugin type`


## Initialization mechanism

Users can easily find a plugin type and initialize an empty plugin instance according to the previous registration mechanism. However, users often need an initialized plugin rather than a empty plugin. So we define the initialization mechanism in
Plugin structure.

In the initialization mechanism, the plugin category(interface) and the init config is required. Initialize processing is like the following.

1. Find the plugin name in the input config.
2. Find plugin type according to the plugin category(interface) and the plugin name.
3. Create an empty plugin.
4. Initialize the plugin according to the config.
5. do some callback processing after initialized.


## Plugin usage in Satellite
Not only does the extension mechanism depend on Plugin, but the core modules also depend on Plugin in Satellite. We'll illustrate this with an example.

- Core Module interface: module.Service
    - the specific plugin: module.Gatherer
    - the specific plugin: module.Processor
    - the specific plugin: module.Sender
    - the specific plugin: module.ClientManager
- Extension: 
    - Collector interface 
        - the specific plugin: segment-receiver
    - Queue interface
        - the specific plugin: mmap-queue
    - Filter interface
        - the specific plugin: sampling-filter
    - Client interface
        - the specific plugin: gRpc-client

Extension plugins constitute the specific module plugin, such as the segment-receiver plugin is used in the Gatherer plugin. And the module plugins constitute Satellite.
