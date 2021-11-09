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

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

// the global plugin registry
var (
	Reg map[reflect.Type]map[string]reflect.Value
)

func init() {
	Reg = make(map[reflect.Type]map[string]reflect.Value)
}

// RegisterPluginCategory register the RegInfo to the global type registry.
func RegisterPluginCategory(pluginType reflect.Type) {
	Reg[pluginType] = map[string]reflect.Value{}
}

// RegisterPlugin registers the pluginType as plugin.
// If the plugin is a pointer receiver, please pass a pointer. Otherwise, please pass a value.
func RegisterPlugin(plugin Plugin) {
	v := reflect.ValueOf(plugin)
	success := false
	for pCategory, pReg := range Reg {
		if v.Type().Implements(pCategory) {
			pReg[plugin.Name()] = v
			log.Logger.WithFields(logrus.Fields{
				"category":    v.Type().String(),
				"plugin_name": plugin.Name(),
			}).Debug("register plugin success")
			success = true
		}
	}
	if !success {
		log.Logger.WithFields(logrus.Fields{
			"category":    v.Type().String(),
			"plugin_name": plugin.Name(),
		}).Error("plugin is not allowed to register")
	}
}

// Get an initialized specific plugin according to the pluginCategory and config.
func Get(category reflect.Type, cfg Config) Plugin {
	pluginName := nameFinder(cfg)
	value, ok := Reg[category][pluginName]
	if !ok {
		panic(fmt.Errorf("cannot find %s plugin, and the category of plugin is %s", pluginName, category))
	}
	t := value.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	plugin := reflect.New(t).Interface().(Plugin)
	Initializing(plugin, cfg)
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

// Initializing initialize the fields by fields mapping.
func Initializing(plugin Plugin, cfg Config) {
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
	cf := reflect.ValueOf(plugin).Elem().FieldByName(config.CommonFieldsName)
	if !cf.IsValid() {
		panic(fmt.Errorf("%s plugin must have a field named CommonField", plugin.Name()))
	}
	for i := 0; i < cf.NumField(); i++ {
		tagVal := cf.Type().Field(i).Tag.Get(config.TagName)
		if tagVal != "" {
			if val := cfg[strings.ToLower(config.CommonFieldsName)+"_"+tagVal]; val != nil {
				cf.Field(i).Set(reflect.ValueOf(val))
			}
		}
	}
}
