```
// The following graph illustrates the relationship between different plugin interface in api package.
//
//
//                   Gatherer                                Processor
//       -------------------------------      -------------------------------------------
//      |  -----------       ---------   |   |  -----------                 -----------  |
//      | | Collector | ==> |  Queue   | |==>| |  Filter   | ==>  ...  ==> |  Filter   | |
//      | | (Parser)  |     | Mem/File | |   |  -----------                 -----------  |
//      |  -----------       ----------  |   |      ||                          ||       |
//       --------------------------------    |      \/	                        \/       |
//                                           |  ---------------------------------------  |
//                                           | |             OutputEventContext        | |
//                                           |  ---------------------------------------  |
//                                            -------------------------------------------      -----------------
//                                             ||                                       ----->| Sharing Client  |
//                                             \/              Sender                  |       -----------------
//                                             ----------------------------------------|-
//                                            |  ---       ---                         | |
//                                            | | B |     | D |     -----------------  | |
//                                            | | A |     | I |    |Segment Forwarder|-| |
//                                            | | T |     | S |    |    (Fallbacker) | | |
//                                            | | C |     | P |     -----------------  | |
//                                            | | H |  => | A |                        | | ===> Kakfa/OAP
//                                            | | B |     | T | =>        ......       | |
//                                            | | U |     | C |                        | |
//                                            | | F |     | H |     -----------------  | |
//                                            | | F |     | E |    | Meter  Forwarder|-| |
//                                            | | E |     | R |    |     (Fallbacker | | |
//                                            | | R |     |   |     -----------------  | |
//                                            |  ---       ---                         | |
//                                             ----------------------------------------
//
//
// 1. The Collector plugin would fetch or receive the input data.
// 2. The Parser plugin would parse the input data to SerializableEvent that is supported
//    to be stored in Queue.
// 3. The Queue plugin stores the SerializableEvent. However, whether serializing depends on
//    the Queue implements. For example, the serialization is unnecessary when using a Memory
//    Queue. Once an event is pulled by the consumer of Queue, the event will be processed by
//    the filters in Processor.
// 4. The Filter plugin would process the event to create a new event. Next, the event is passed
//    to the next filter to do the same things until the whole filters are performed. All created
//    events would be stored in the OutputEventContext. However, only the events labeled with
//    RemoteEvent type would be forwarded by Forwarder.
// 5. After processing, the events in OutputEventContext would be stored in the BatchBuffer. When
//    the timer is triggered or the capacity limit is reached, the events in BatchBuffer would be
//    partitioned by EventType and sent to the different Forwarders, such as Segment Forwarder and
//    Meter Forwarder.
// 6. The Follower in different Senders would share with the remote client to avoid make duplicate
//    connections and have the same Fallbacker(FallBack strategy) to process data. When all
//    forwarders send success or process success in Fallbacker, the dispatcher would also ack the
//    batch is a success.

// ============================================================================================
//
// There are four stages in the lifecycle of Satellite plugins, which are the initial phase,
// preparing phase, running phase, and closing phase. In the running phase, each plugin has
// its own interface definition. However, the other three phases have to be defined uniformly.
```