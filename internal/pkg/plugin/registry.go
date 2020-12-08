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

// the global plugin registry
var (
	lock              sync.Mutex
	reg               map[reflect.Type]map[string]reflect.Value
	initFuncReg       map[reflect.Type]InitializingFunc
	callbackFuncReg   map[reflect.Type]CallbackFunc
	nameFinderFuncReg map[reflect.Type]NameFinderFunc
)

func init() {
	reg = make(map[reflect.Type]map[string]reflect.Value)
	initFuncReg = make(map[reflect.Type]InitializingFunc)
	callbackFuncReg = make(map[reflect.Type]CallbackFunc)
	nameFinderFuncReg = make(map[reflect.Type]NameFinderFunc)
}

// RegisterCategory register new plugin category with default InitializingFunc.
// required:
// pluginCategory: the plugin interface type.
// Optional:
// n: the plugin name finder,and the default value is defaultNameFinder
// i, the plugin initializer, and the default value is defaultInitializing
// c, the plugin initializer callback func, and the default value is defaultCallBack
func RegisterPluginCategory(pluginCategory reflect.Type, n NameFinderFunc, i InitializingFunc, c CallbackFunc) {
	lock.Lock()
	defer lock.Unlock()
	reg[pluginCategory] = map[string]reflect.Value{}

	if n == nil {
		nameFinderFuncReg[pluginCategory] = defaultNameFinder
	} else {
		nameFinderFuncReg[pluginCategory] = n
	}
	if i == nil {
		initFuncReg[pluginCategory] = defaultInitializing
	} else {
		initFuncReg[pluginCategory] = i
	}
	if c == nil {
		callbackFuncReg[pluginCategory] = defaultCallBack
	} else {
		callbackFuncReg[pluginCategory] = c
	}
}

// RegisterPlugin registers the pluginType as plugin.
// If the plugin is a pointer receiver, please pass a pointer. Otherwise, please pass a value.
func RegisterPlugin(plugin Plugin) {
	lock.Lock()
	defer lock.Unlock()
	v := reflect.ValueOf(plugin)
	success := false
	for pCategory, pReg := range reg {
		if v.Type().Implements(pCategory) {
			pReg[plugin.Name()] = v
			fmt.Printf("register %s %s successfully ", plugin.Name(), v.Type().String())
			success = true
		}
	}
	if !success {
		fmt.Printf("this type of %s is not supported to register : %s", plugin.Name(), v.Type().String())
	}
}

// Get an initialized specific plugin according to the pluginCategory and pluginName.
func Get(category reflect.Type, cfg interface{}) Plugin {
	lock.Lock()
	defer lock.Unlock()
	pluginName := nameFinderFuncReg[category](cfg)
	value, ok := reg[category][pluginName]
	if !ok {
		panic(fmt.Errorf("cannot find %s plugin, and the category of plugin is %s", pluginName, category))
	}
	t := value.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	plugin := reflect.New(t).Interface().(Plugin)
	initFuncReg[category](plugin, cfg)
	callbackFuncReg[category](plugin)
	return plugin
}
