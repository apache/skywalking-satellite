// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package api

import "io"

// The following comments is to illustrate the relationship between different plugin interface in api package.
//
//
//                   Gatherer                                Processor
//       -------------------------------      -------------------------------------------
//      | | -----------        --------- |   |  -----------                 -----------  |
//      | | Collector | ==> |  Queue   | |==>| |  Filter   | ==>  ...  ==> |  Filter   | |
//      | | (Parser)  |     | Mem/File | |   |  -----------                 -----------  |
//      |  -----------       ---------   |   |      ||                          ||       |
//       --------------------------------    |      \/	                        \/       |
//                                           |  ---------------------------------------  |
//                                           | |             OutputEventContext        | |
//                                           |  ---------------------------------------  |
//                                            -------------------------------------------
//                             									   ||
//                                                                 \/
//                                            -------------------------------------------
//                                           |                                   ||      |
//                                           |                                   \/      |
//                                           |  -------------------       -------------  |
//                                           | | BatchOutputEvents | <== | BatchBuffer | |
//                                           |  -------------------       -------------  |
//                                   Sender  |             ||                            | ==> Kafka/OAP
//                                           |             \/                            |
//                                           |  -------------------                      |
//                                           | |     Forwarder     |                     |
//                                           |  -------------------                      |
//                                           |                                           |
//                                            -------------------------------------------
//
// 1. The Collector plugin would fetch or receive the input data.
// 2. The Parser plugin would parse the input data to InputEvent.
//    If the event needs output, please tag it by the IsOutput
//    method.
// 3. The Queue plugin would store the InputEvent. But different
//    Queue would use different ways to store data, such as store
//    bytes by serialization or keep original.
// 4. The Filter plugin would pull the event from the Queue and
//    process the event to create a new event. Next, the event is
//    passed to the next filter to do the same things until the
//    whole processor are performed. Similar to above, if any
//    events need output, please mark. The events would be stored
//    in the OutputEventContext. When the processing is finished,
//    the OutputEventContext would be passed to the BatchBuffer.
// 5. When BatchBuffer reaches its maximum capacity, the
//    OutputEventContexts would be partitioned by event name and
//    convert to BatchOutputEvents.
// 6. The Follower would be ordered to send each partition in
//    BatchOutputEvents in different ways, such as different gRPC
//    endpoints.

// ComponentPlugin is an interface to initialize the specific plugin.
type ComponentPlugin interface {
	io.Closer

	// Init initialize the specific plugin and would return error when the configuration is error.
	Init() error

	// Run triggers the specific plugin to work, such as build connection.
	Run()
}
