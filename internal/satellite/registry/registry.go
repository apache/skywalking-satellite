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
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/api"
)

// All plugins is wrote in ./plugins dir. The plugin type would be as the next level dirs,
// such as collector, client, or queue. And the 3rd level is the plugin name, that is also
// used as key in pluginRegistry.

// reg is the global plugin registry
var reg map[reflect.Type]map[string]interface{}
var lock sync.Mutex

// Supported plugin types
var (
	collectorType  = reflect.TypeOf((*api.Collector)(nil)).Elem()
	queueType      = reflect.TypeOf((*api.Queue)(nil)).Elem()
	filterType     = reflect.TypeOf((*api.Filter)(nil)).Elem()
	forwarderType  = reflect.TypeOf((*api.Forwarder)(nil)).Elem()
	parserType     = reflect.TypeOf((*api.Parser)(nil)).Elem()
	clientType     = reflect.TypeOf((*api.Client)(nil)).Elem()
	fallbackerType = reflect.TypeOf((*api.Fallbacker)(nil)).Elem()
)

func init() {
	reg = make(map[reflect.Type]map[string]interface{})
	reg[collectorType] = make(map[string]interface{})
	reg[queueType] = make(map[string]interface{})
	reg[filterType] = make(map[string]interface{})
	reg[forwarderType] = make(map[string]interface{})
	reg[parserType] = make(map[string]interface{})
	reg[clientType] = make(map[string]interface{})
	reg[fallbackerType] = make(map[string]interface{})
}

// RegisterPlugin registers the pluginType as plugin.
func RegisterPlugin(pluginName string, plugin interface{}) {
	lock.Lock()
	defer lock.Unlock()
	t := reflect.TypeOf(plugin)
	success := false
	for pType, pReg := range reg {
		if t.Implements(pType) {
			pReg[pluginName] = plugin
			fmt.Printf("register %s %s successfully ", pluginName, t.String())
			success = true
			break
		}
	}
	if !success {
		fmt.Printf("this type of %s is not supported to register : %s", pluginName, t.String())
	}
}

func GetCollector(pluginName string) api.Collector {
	plugin := reg[collectorType][pluginName]
	if plugin != nil {
		return plugin.(api.Collector)
	}
	return nil
}

func GetQueue(pluginName string) api.Queue {
	plugin := reg[queueType][pluginName]
	if plugin != nil {
		return plugin.(api.Queue)
	}
	return nil
}

func GetFilter(pluginName string) api.Filter {
	plugin := reg[filterType][pluginName]
	if plugin != nil {
		return plugin.(api.Filter)
	}
	return nil
}

func GetForwarder(pluginName string) api.Forwarder {
	plugin := reg[forwarderType][pluginName]
	if plugin != nil {
		return plugin.(api.Forwarder)
	}
	return nil
}
func GetParser(pluginName string) api.Parser {
	plugin := reg[parserType][pluginName]
	if plugin != nil {
		return plugin.(api.Parser)
	}
	return nil
}

func GetFallbacker(pluginName string) api.Fallbacker {
	plugin := reg[fallbackerType][pluginName]
	if plugin != nil {
		return plugin.(api.Fallbacker)
	}
	return nil
}

func GetClient(pluginName string) api.Client {
	plugin := reg[clientType][pluginName]
	if plugin != nil {
		return plugin.(api.Client)
	}
	return nil
}
