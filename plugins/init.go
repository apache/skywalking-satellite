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
	"github.com/apache/skywalking-satellite/plugins/client"
	"github.com/apache/skywalking-satellite/plugins/fallbacker"
	fetcher "github.com/apache/skywalking-satellite/plugins/fetcher"
	filter "github.com/apache/skywalking-satellite/plugins/filter/api"
	"github.com/apache/skywalking-satellite/plugins/forwarder"
	parser "github.com/apache/skywalking-satellite/plugins/parser/api"
	"github.com/apache/skywalking-satellite/plugins/queue"
	"github.com/apache/skywalking-satellite/plugins/receiver"
	"github.com/apache/skywalking-satellite/plugins/server"
)

// RegisterPlugins register the whole supported plugin category and plugin types to the registry.
func RegisterPlugins() {
	// plugins
	filter.RegisterFilterPlugins()
	forwarder.RegisterForwarderPlugins()
	parser.RegisterParserPlugins()
	queue.RegisterQueuePlugins()
	receiver.RegisterReceiverPlugins()
	fetcher.RegisterFetcherPlugins()
	fallbacker.RegisterFallbackerPlugins()
	// sharing plugins
	server.RegisterServerPlugins()
	client.RegisterClientPlugins()
}
