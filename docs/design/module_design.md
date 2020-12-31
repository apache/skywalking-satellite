# Module Design
## Pipe
The pipe is an isolation concept in Satellite. 
Each pipe has one pipeline to process the telemetry data(metrics/traces/logs). Two pipes are not sharing data.

```
                            Satellite
 ---------------------------------------------------------------------
|            -------------------------------------------              |
|           |                 Pipe                      |             |
|            -------------------------------------------              |
|            -------------------------------------------              |
|           |                 Pipe                      |             |
|            -------------------------------------------              |
|            -------------------------------------------              |
|           |                 Pipe                      |             |
|            -------------------------------------------              |
 ---------------------------------------------------------------------
```
## Modules
There are 3 modules in one pipe, which are Gatherer, Processor, and Sender.

- The Gatherer module is responsible for fetching or receiving data and pushing the data to Queue. So there are 2 kinds of Gatherer, which are ReceiverGatherer and FetcherGatherer.
- The Processor module is responsible for reading data from the queue and processing data by a series of filter chains.
- The Sender module is responsible for async processing and forwarding the data to the external services in the batch mode. After sending success, Sender would also acknowledge the offset of Queue in Gatherer.

```
                            Pipe
 --------------------------------------------------------------------
|            ----------      -----------      --------               |
|           | Gatherer | => | Processor | => | Sender |              |                          
|            ----------      -----------      --------               |
 --------------------------------------------------------------------
```

## Plugins

Plugin is the minimal components in the module. Sateliite has 2 plugin catalogs, which are sharing plugins and normal plugins.

- a sharing plugin instance could be sharing with multiple modules in the different pipes.
- a normal plugin instance is only be used in a fixed module of the fixed pipes.

### Sharing plugin
Nowadays, there are 2 kinds of sharing plugins in Satellite, which are server plugins and client plugins. The reason why they are sharing plugins is to reduce the resource cost in connection. Server plugins are sharing with the ReceiverGatherer modules in the different pipes to receive the external requests. And the client plugins is sharing with the Sender modules in the different pipes to connect with external services, such as Kafka and OAP.

```
           Sharing Server                      Sharing Client
 --------------------------------------------------------------------
|       ------------------      -----------      --------            |
|      | ReceiverGatherer | => | Processor | => | Sender |           |                          
|       ------------------      -----------      --------            |
 --------------------------------------------------------------------
 --------------------------------------------------------------------
|       ------------------      -----------      --------            |
|      | ReceiverGatherer | => | Processor | => | Sender |           |                          
|       ------------------      -----------      --------            |
 --------------------------------------------------------------------
 --------------------------------------------------------------------
|       ------------------      -----------      --------            |
|      | ReceiverGatherer | => | Processor | => | Sender |           |                          
|       ------------------      -----------      --------            |
 --------------------------------------------------------------------
```

### Normal plugin
There are 7 kinds of normal plugins in Satellite, which are Receiver, Fetcher, Queue, Parser, Filter, Forwarder, and Fallbacker.

- Receiver: receives the input APM data from the request.
- Fetcher: fetch the APM data by fetching.
- Queue: store the APM data to ensure the data stability.
- Parser: supports some ways to parse data, such parse a csv file.
- Filter: processes the APM data.
- Forwarder: forwards the APM data to the external receiver, such as Kafka and OAP.
- Fallbacker: supports some fallback strategies, such as timer retry strategy.

```

                   Gatherer                                Processor
       -------------------------------      -------------------------------------------
      |  -----------       ---------   |   |  -----------                 -----------  |
      | | Receiver  | ==> |  Queue   | |==>| |  Filter   | ==>  ...  ==> |  Filter   | |
      | | /Fetcher  |     | Mem/File | |   |  -----------                 -----------  |
      |  -----------       ----------  |   |      ||                          ||       |
       --------------------------------    |      \/	                      \/       |
                                           |  ---------------------------------------  |
                                           | |             OutputEventContext        | |
                                           |  ---------------------------------------  |
                                            -------------------------------------------     
                                             ||                                      
                                             \/              Sender                  
                                             ------------------------------------------
                                            |  ---       ---                           |
                                            | | B |     | D |     -----------------    |
                                            | | A |     | I |    |Segment Forwarder|   |
                                            | | T |     | S |    |    (Fallbacker) |   |
                                            | | C |     | P |     -----------------    |
                                            | | H |  => | A |                          | ===> Kakfa/OAP
                                            | | B |     | T | =>        ......         |
                                            | | U |     | C |                          |
                                            | | F |     | H |     -----------------    |
                                            | | F |     | E |    | Meter  Forwarder|   |
                                            | | E |     | R |    |     (Fallbacker |   |
                                            | | R |     |   |     -----------------    |
                                            |  ---       ---                           |
                                             ------------------------------------------


 1. The Fetcher/Receiver plugin would fetch or receive the input data.
 2. The Parser plugin would parse the input data to SerializableEvent that is supported
    to be stored in Queue.
 3. The Queue plugin stores the SerializableEvent. However, whether serializing depends on
    the Queue implements. For example, the serialization is unnecessary when using a Memory
    Queue. Once an event is pulled by the consumer of Queue, the event will be processed by
    the filters in Processor.
 4. The Filter plugin would process the event to create a new event. Next, the event is passed
    to the next filter to do the same things until the whole filters are performed. All created
    events would be stored in the OutputEventContext. However, only the events labeled with
    RemoteEvent type would be forwarded by Forwarder.
 5. After processing, the events in OutputEventContext would be stored in the BatchBuffer. When
    the timer is triggered or the capacity limit is reached, the events in BatchBuffer would be
    partitioned by EventType and sent to the different Forwarders, such as Segment Forwarder and
    Meter Forwarder.
 6. The Follower in different Senders would share with the remote client to avoid make duplicate
    connections and have the same Fallbacker(FallBack strategy) to process data. When all
    forwarders send success or process success in Fallbacker, the dispatcher would also ack the
    batch is a success.
 ============================================================================================
```