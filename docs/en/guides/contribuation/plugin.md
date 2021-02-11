# How to write a new plugin?
If you want to add a custom plugin in SkyWalking Satellite, the following contents would guide you.
Let's use memory-queue as an example of how to write a plugin.

1. Choose the plugin category. As the memory-queue is a queue, the plugin should be written in the **skywalking-satellite/plugins/queue** directory. So we create a new directory called memory as the plugin codes space.  

2. Implement the interface in the **skywalking-satellite/plugins/queue/api**. Each plugin has 3 common methods, which are Name(), Description(), DefaultConfig().
    - Name() returns the unique name in the plugin category.
    - Description() returns the description of the plugin, which would be used to generate the plugin documentation.
    - DefaultConfig() returns the default plugin config with yaml pattern, which would be used as the default value in the plugin struct and to generate the plugin documentation.
    ```go
    type Queue struct {
    	config.CommonFields
    	// config
    	EventBufferSize int `mapstructure:"event_buffer_size"` // The maximum buffer event size.
    
    	// components
    	buffer *goconcurrentqueue.FixedFIFO
    }
    
    func (q *Queue) Name() string {
    	return Name
    }
    
    func (q *Queue) Description() string {
    	return "this is a memory queue to buffer the input event."
    }
    
    func (q *Queue) DefaultConfig() string {
    	return `
    # The maximum buffer event size.
    event_buffer_size: 5000
    ```
   
3. Add [unit test](../test/test.md).
4. Generate the plugin docs.
```shell script
make check
```



