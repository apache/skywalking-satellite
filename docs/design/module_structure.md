# Module structure

Module is the core workers in Satellite. Module is constituted by the specific extension plugins.
There are four modules in Satellite, which is ClientManager, Gatherer, Sender, and Processor.

Responsibilities:

- ClientManager: Maintain connection and monitor connection status
- Sender: Sender data to the external services, such as Kafka and OAP
- Gatherer: Gather the APM data from the other systems, such as fetch prometheus metrics.
- Processor: Data processing to create new metrics data.

LifeCycles:

- Prepare: Prepare phase is to do some preparation works, such as make the connection with external services.
- Boot: Boot phase is to start the current module until receives a close signal.
- ShutDown: ShutDown phase is to close the used resources.

