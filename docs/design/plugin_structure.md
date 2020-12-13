# plugin structure
`Plugin is a common concept for Satellite, which is in all externsion plugins.`

## Registration mechanism

The Plugin registration mechanism in Satellite is similar to the SPI registration mechanism of Java. 
Plugin registration mechanism supports to register an interface and its implementation, that means different interfaces have different registration spaces.
We can easily find the type of a specific plugin according to the interface and the plugin name and initialize it according to the type.

structure:
- code: `map[reflect.Type]map[string]reflect.Value`
- meaning: `map[interface type]map[plugin name] plugin type`


## Initialization mechanism

Users can easily find a plugin type and initialize an empty plugin instance according to the previous registration mechanism. For setting up the configuration of the extension convenience, we define the initialization mechanism in Plugin structure.

In the initialization mechanism, `the plugin category(interface)` and `the init config is required`.
 
Initialize processing is like the following.

1. Find the plugin name in the input config according to the fixed key `plugin_name`.
2. Find plugin type according to the plugin category(interface) and the plugin name.
3. Create an empty plugin.
4. Initialize the plugin according to the merged config, which is created by the input config and the default config.



## Plugin usage in Satellite
Nowadays, the numbers of the Plugin categories is 2. One is the [sharing Plugin](module_design.md), and another is the other [normal Plugin](module_design.md).

- Extension Plugins: 
    - sharing plugins
        - Server Plugin
        - Client Plugin
    - normal plugins
        - Receiver Plugin
        - Fetcher Plugin
        - Parser Plugin
        - Queue Plugin
        - Filter Plugin
        - Fallbacker Plugin
        - Forwarder Plugin
