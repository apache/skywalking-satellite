# Module structure

## Overview
Module is the core workers in Satellite. Module is constituted by the specific extension plugins.
There are 3 modules in one namespace, which are Gatherer, Processor, and Sender.

- The Gatherer module is responsible for fetching or receiving data and pushing the data to Queue. So there are 2 kinds of Gatherer, which are ReceiverGatherer and FetcherGatherer.
- The Processor module is responsible for reading data from the queue and processing data by a series of filter chains.
- The Sender module is responsible for async processing and forwarding the data to the external services in the batch mode. After sending success, Sender would also acknowledge the offset of Queue in Gatherer.

```
                            Namespace
 --------------------------------------------------------------------
|            ----------      -----------      --------               |
|           | Gatherer | => | Processor | => | Sender |              |                          
|            ----------      -----------      --------               |
 --------------------------------------------------------------------
```

## LifeCycle

- Prepare: Prepare phase is to do some preparation works, such as register the client status listener to the client in ReceiverGatherer.
- Boot: Boot phase is to start the current module until receives a close signal.
- ShutDown: ShutDown phase is to close the used resources.

