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

package plugin

import (
	"fmt"
	"reflect"
	"sync"
)

// All plugins is wrote in ./plugins dir. The plugin type would be as the next level dirs,
// such as collector, client, or queue. And the 3rd level is the plugin name, that is also
// used as key in pluginRegistry.

// reg is the global plugin registry
var (
	reg  map[reflect.Type]map[string]reflect.Value
	lock sync.Mutex
)

func init() {
	reg = make(map[reflect.Type]map[string]reflect.Value)
}

// Add new plugin category. The different plugin category could have same plugin names.
func AddPluginCategory(pluginCategory reflect.Type) {
	lock.Lock()
	defer lock.Unlock()
	reg[pluginCategory] = map[string]reflect.Value{}
}

// RegisterPlugin registers the pluginType as plugin.
// If the plugin is a pointer receiver, please pass a pointer. Otherwise, please pass a value.
func RegisterPlugin(pluginName string, plugin interface{}) {
	lock.Lock()
	defer lock.Unlock()
	v := reflect.ValueOf(plugin)
	success := false
	for pCategory, pReg := range reg {
		if v.Type().Implements(pCategory) {
			pReg[pluginName] = v
			fmt.Printf("register %s %s successfully ", pluginName, v.Type().String())
			success = true
		}
	}
	if !success {
		fmt.Printf("this type of %s is not supported to register : %s", pluginName, v.Type().String())
	}
}

// Get the specific plugin according to the pluginCategory and pluginName.
func Get(pluginCategory reflect.Type, pluginName string, config map[string]interface{}) Plugin {
	value, ok := reg[pluginCategory][pluginName]
	if !ok {
		panic(fmt.Errorf("cannot find %s plugin, and the category of plugin is %s", pluginName, pluginCategory))
	}
	t := value.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	plugin := reflect.New(t).Interface().(Plugin)
	plugin.InitPlugin(config)
	return plugin
}
