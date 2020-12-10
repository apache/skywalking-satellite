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
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// the global plugin registry
var (
	lock sync.Mutex
	reg  map[reflect.Type]map[string]reflect.Value
)

func init() {
	reg = make(map[reflect.Type]map[string]reflect.Value)
}

// RegisterPluginCategory register the RegInfo to the global type registry.
func RegisterPluginCategory(pluginType reflect.Type) {
	lock.Lock()
	defer lock.Unlock()
	reg[pluginType] = map[string]reflect.Value{}
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

// Get an initialized specific plugin according to the pluginCategory and config.
func Get(category reflect.Type, cfg Config) Plugin {
	lock.Lock()
	defer lock.Unlock()
	pluginName := nameFinder(cfg)
	value, ok := reg[category][pluginName]
	if !ok {
		panic(fmt.Errorf("cannot find %s plugin, and the category of plugin is %s", pluginName, category))
	}
	t := value.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	plugin := reflect.New(t).Interface().(Plugin)
	initializing(plugin, cfg)
	return plugin
}

// nameFinder is used to get the plugin name in Config.
func nameFinder(cfg interface{}) string {
	c, ok := cfg.(Config)
	if !ok {
		panic(fmt.Errorf("nameFinder only supports Config"))
	}
	name, ok := c[NameField]
	if !ok {
		panic(fmt.Errorf("%s is requeired in Config", NameField))
	}
	return name.(string)
}

// initializing initialize the fields by fields mapping.
func initializing(plugin Plugin, cfg Config) {
	v := viper.New()
	v.SetConfigType("yaml")
	if plugin.DefaultConfig() != "" {
		if err := v.ReadConfig(strings.NewReader(plugin.DefaultConfig())); err != nil {
			panic(fmt.Errorf("cannot read default config in the plugin: %s, the error is %v", plugin.Name(), err))
		}
	}
	if err := v.MergeConfigMap(cfg); err != nil {
		panic(fmt.Errorf("%s plugin cannot merge the custom configuration, the error is %v", plugin.Name(), err))
	}
	if err := v.Unmarshal(plugin); err != nil {
		panic(fmt.Errorf("cannot inject  the config to the %s plugin, the error is %v", plugin.Name(), err))
	}
}
