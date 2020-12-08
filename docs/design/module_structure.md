# Module structure

Each module is a plugin in Satellite. According to the extension mechanism in Plugin system, `module.Service` supports DI to decouple the dependencies of different modules.

## Module phases
`module.Service` has 4 phases in the life cycle, which are Init, Prepare, boot and shutDown.

- Init: Init phase is running in Plugin system to initialize a plugin, and register it to the module container.
- Prepare: Prepare phase is to do some preparation works, such as make the connection with external services. And do dependency injection depends on the module container.
- Boot: Boot phase is to start the current module until receives a close signal.
- ShutDown: ShutDown phase is to close the used resources.
