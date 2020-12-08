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
	"strings"

	"github.com/spf13/viper"
)

// DefaultPluginNameField is a required field in DefaultConfig.
const DefaultPluginNameField = "plugin_name"

// DefaultInitializingPlugin defines the plugins initialized by defaultInitializing.
type DefaultInitializingPlugin interface {
	Plugin
	// DefaultConfig returns the default config, that is a YAML pattern.
	DefaultConfig() string
}

// DefaultConfig is used to initialize the DefaultInitializingPlugin.
type DefaultConfig map[string]interface{}

// defaultNameFinder is used to get the plugin name in DefaultConfig.
func defaultNameFinder(cfg interface{}) string {
	c, ok := cfg.(DefaultConfig)
	if !ok {
		panic(fmt.Errorf("defaultNameFinder only supports DefaultConfig"))
	}
	name, ok := c[DefaultPluginNameField]
	if !ok {
		panic(fmt.Errorf("%s is requeired in DefaultConfig", DefaultPluginNameField))
	}
	return name.(string)
}

// defaultInitializing initialize the fields by fields mapping.
func defaultInitializing(plugin Plugin, cfg interface{}) {
	c, ok := cfg.(DefaultConfig)
	if !ok {
		panic(fmt.Errorf("%s plugin is a DefaultInitializingPlugin, but the type of configuration is illegal", plugin.Name()))
	}
	v := viper.New()
	v.SetConfigType("yaml")
	p := plugin.(DefaultInitializingPlugin)
	if p.DefaultConfig() != "" {
		if err := v.ReadConfig(strings.NewReader(p.DefaultConfig())); err != nil {
			panic(fmt.Errorf("cannot read default config in the plugin: %s, the error is %v", plugin.Name(), err))
		}
	}
	if err := v.MergeConfigMap(c); err != nil {
		panic(fmt.Errorf("%s plugin cannot merge the custom configuration, the error is %v", plugin.Name(), err))
	}
	if err := v.Unmarshal(plugin); err != nil {
		panic(fmt.Errorf("cannot inject  the config to the %s plugin, the error is %v", plugin.Name(), err))
	}
}

// defaultCallBack does nothing.
func defaultCallBack(plugin Plugin) {}
