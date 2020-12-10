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

package plugins

import (
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	collector "github.com/apache/skywalking-satellite/plugins/collector/api"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	filter "github.com/apache/skywalking-satellite/plugins/filter/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	parser "github.com/apache/skywalking-satellite/plugins/parser/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
)

func RegisterPlugins() {
	filter.RegisterFilterPlugins()
	forwarder.RegisterForwarderPlugins()
	parser.RegisterParserPlugins()
	queue.RegisterQueuePlugins()
	client.RegisterClientPlugins()
	collector.RegisterCollectorPlugins()
	fallbacker.RegisterFallbackerPlugins()
}
