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
	"reflect"
	"testing"
)

type DemoCategory interface {
	DefaultInitializingPlugin
	Say() string
}

type DemoPlugin struct {
	Organization string `mapstructure:"organization"`
	Project      string `mapstructure:"project"`
}

func (d *DemoPlugin) Say() string {
	return d.Organization + ":" + d.Project
}

func (d *DemoPlugin) Name() string {
	return "demoplugin"
}

func (d *DemoPlugin) Description() string {
	return "this is just a demo"
}

func (d *DemoPlugin) DefaultConfig() string {
	return `
organization: "ASF"
project: "skywalking-satellite"
`
}

func TestPlugin(t *testing.T) {
	tests := []struct {
		name string
		args DefaultConfig
		want *DemoPlugin
	}{
		{
			name: "test1",
			args: DefaultConfig{
				"plugin_name":  "demoplugin",
				"organization": "CNCF",
				"project":      "Fluentd",
			},
			want: &DemoPlugin{
				Organization: "CNCF",
				Project:      "Fluentd",
			},
		},
		{
			name: "demoplugin",
			args: DefaultConfig{
				"plugin_name": "demoplugin",
			},
			want: &DemoPlugin{
				Organization: "ASF",
				Project:      "skywalking-satellite",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if i := recover(); i != nil {
					t.Errorf("the plugin %s is not exist", "demoplugin")
				}
			}()
			plugin := Get(reflect.TypeOf((*DemoCategory)(nil)).Elem(), tt.args)
			if !reflect.DeepEqual(plugin, tt.want) {
				t.Errorf("Format() got = %v, want %v", plugin, tt.want)
			}
		})
	}
}

func init() {
	RegisterPluginCategory(reflect.TypeOf((*DemoCategory)(nil)).Elem(), nil, nil, nil)
	RegisterPlugin(&DemoPlugin{})
}
