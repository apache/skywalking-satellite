# How do we run Satellite on Windows?

Windows is not good supported because [some features](../guides/compile/compile.md) is not adaptive on the Windows. Let's take the mmap component as an example to talk how to solve the problem.

1. Remove package dependency. Let's open the SkyWalking-Satellite/plugins/queue/queue_repository.go, and delete the pointer line.
```go
// RegisterQueuePlugins register the used queue plugins.
func RegisterQueuePlugins() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Queue)(nil)).Elem())
	queues := []api.Queue{
		// Please register the queue plugins at here.
		new(memory.Queue),
		new(mmap.Queue), <=====Delete the line.
	}
	for _, q := range queues {
		plugin.RegisterPlugin(q)
	}
}
```
2. Append the windows platform to the build script in the makefile.

```
.PHONY: build
build: deps linux darwin windows
```
3. Execute the build command.
```
make build.
```