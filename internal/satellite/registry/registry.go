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

package registry

import (
	"fmt"
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/api"
)

// The creator registry.
// All plugins is wrote in ./plugins dir. The plugin type would be as the next level dirs,
// such as collector, client, or queue. And the 3rd level is the plugin name, that is also
// used as key in pluginRegistry.
type pluginRegistry struct {
	collectorRegistry  map[string]api.Collector
	queueRegistry      map[string]api.Queue
	filterRegistry     map[string]api.Filter
	forwarderRegistry  map[string]api.Forwarder
	parserRegistry     map[string]api.Parser
	clientRegistry     map[string]api.Client
	fallbackerRegistry map[string]api.Fallbacker
}

// reg is the global plugin registry
var reg *pluginRegistry

// Plugin type.
var (
	collectorType  = reflect.TypeOf((*api.Collector)(nil)).Elem()
	queueType      = reflect.TypeOf((*api.Queue)(nil)).Elem()
	filterType     = reflect.TypeOf((*api.Filter)(nil)).Elem()
	forwardType    = reflect.TypeOf((*api.Forwarder)(nil)).Elem()
	parserType     = reflect.TypeOf((*api.Parser)(nil)).Elem()
	clientType     = reflect.TypeOf((*api.Client)(nil)).Elem()
	fallbackerType = reflect.TypeOf((*api.Fallbacker)(nil)).Elem()
)

func init() {
	reg = &pluginRegistry{}
	reg.collectorRegistry = make(map[string]api.Collector)
	reg.queueRegistry = make(map[string]api.Queue)
	reg.filterRegistry = make(map[string]api.Filter)
	reg.forwarderRegistry = make(map[string]api.Forwarder)
	reg.parserRegistry = make(map[string]api.Parser)
	reg.clientRegistry = make(map[string]api.Client)
	reg.fallbackerRegistry = make(map[string]api.Fallbacker)
}

// RegisterPlugin registers the pluginType as plugin.
func RegisterPlugin(pluginType string, plugin interface{}) {
	t := reflect.TypeOf(plugin)
	switch {
	case t.Implements(collectorType):
		fmt.Printf("register %s collector successfully ", t.String())
		reg.collectorRegistry[pluginType] = plugin.(api.Collector)
	case t.Implements(queueType):
		fmt.Printf("register %s queue successfully ", t.String())
		reg.queueRegistry[pluginType] = plugin.(api.Queue)
	case t.Implements(filterType):
		fmt.Printf("register %s filter successfully ", t.String())
		reg.filterRegistry[pluginType] = plugin.(api.Filter)
	case t.Implements(forwardType):
		fmt.Printf("register %s forwarder successfully ", t.String())
		reg.forwarderRegistry[pluginType] = plugin.(api.Forwarder)
	case t.Implements(parserType):
		fmt.Printf("register %s parser successfully ", t.String())
		reg.parserRegistry[pluginType] = plugin.(api.Parser)
	case t.Implements(clientType):
		fmt.Printf("register %s client successfully ", t.String())
		reg.clientRegistry[pluginType] = plugin.(api.Client)
	case t.Implements(fallbackerType):
		fmt.Printf("register %s fallbacker successfully ", t.String())
		reg.fallbackerRegistry[pluginType] = plugin.(api.Fallbacker)
	default:
		fmt.Printf("this type is not supported to register : %s", t.String())
	}
}

func GetCollector(pluginType string) api.Collector {
	return reg.collectorRegistry[pluginType]
}

func GetQueue(pluginType string) api.Queue {
	return reg.queueRegistry[pluginType]
}

func GetFilter(pluginType string) api.Filter {
	return reg.filterRegistry[pluginType]
}

func GetForwarder(pluginType string) api.Forwarder {
	return reg.forwarderRegistry[pluginType]
}
func GetParser(pluginType string) api.Parser {
	return reg.parserRegistry[pluginType]
}

func GetFallbacker(pluginType string) api.Fallbacker {
	return reg.fallbackerRegistry[pluginType]
}

func GetClient(pluginType string) api.Client {
	return reg.clientRegistry[pluginType]
}
