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
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/api"
)

type demoCollector struct {
}

type demoParser struct {
}

type demoQueue struct {
}

type demoFilter struct {
}

type demoForwarder struct {
}

type demoFallbacker struct {
}

type demoClient struct {
}

func init() {

}

func TestRegisterPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoCollector1",
			args: args{
				pluginType: "demoCollector1",
				plugin:     &demoCollector{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoCollector2",
			args: args{
				pluginType: "demoCollector2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetCollector(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterParserPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoParser1",
			args: args{
				pluginType: "demoParser1",
				plugin:     &demoParser{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoParser2",
			args: args{
				pluginType: "demoParser2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetParser(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterQueuePlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoQueue1",
			args: args{
				pluginType: "demoQueue1",
				plugin:     &demoQueue{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoQueue2",
			args: args{
				pluginType: "demoQueue2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetQueue(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterFilterPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoFilter1",
			args: args{
				pluginType: "demoFilter1",
				plugin:     &demoFilter{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoFilter2",
			args: args{
				pluginType: "demoFilter2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetFilter(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterForwarderPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoForwarder1",
			args: args{
				pluginType: "demoForwarder1",
				plugin:     &demoForwarder{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoForwarder2",
			args: args{
				pluginType: "demoForwarder2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetForwarder(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterFallbackerPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoFallbacker1",
			args: args{
				pluginType: "demoFallbacker1",
				plugin:     &demoFallbacker{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoFallbacker2",
			args: args{
				pluginType: "demoFallbacker2",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetFallbacker(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func TestRegisterClientPlugin(t *testing.T) {
	type args struct {
		pluginType string
		plugin     interface{}
	}
	tests := []struct {
		name string
		args args
		want func(interface{}) bool
	}{
		{
			name: "demoClient1",
			args: args{
				pluginType: "demoClient1",
				plugin:     &demoClient{},
			},
			want: func(i interface{}) bool {
				return i != nil
			},
		},
		{
			name: "demoClient2",
			args: args{
				pluginType: "demoClient2",
				plugin:     &demoCollector{},
			},
			want: func(i interface{}) bool {
				return i == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPlugin(tt.args.pluginType, tt.args.plugin)
			if !tt.want(GetClient(tt.args.pluginType)) {
				t.Errorf("the plugin %s is not pass", tt.args.pluginType)
			}
		})
	}
}

func (d demoCollector) InitPlugin() error {
	println("demoCollector init")
	return nil
}

func (d demoCollector) Prepare() {
	println("demoCollector Prepare")
}

func (d demoCollector) Close() error {
	println("demoCollector Close")
	return nil
}

func (d demoCollector) Next() (api.SerializableEvent, error) {
	println("demoCollector Next")
	return nil, nil
}

func (d demoParser) InitPlugin() error {
	println("demoParser init")
	return nil
}

func (d demoParser) ParseBytes(bytes []byte) ([]api.SerializableEvent, error) {
	println("demoParser ParseBytes")
	return nil, nil
}

func (d demoParser) ParseStr(str string) ([]api.SerializableEvent, error) {
	println("demoParser ParseStr")
	return nil, nil
}

func (d demoQueue) InitPlugin() error {
	println("demoQueue init")
	return nil
}

func (d demoQueue) Close() error {
	println("demoQueue Close")
	return nil
}

func (d demoQueue) Publisher() api.QueuePublisher {
	println("demoQueue Publisher")
	return nil
}

func (d demoQueue) Consumer() api.QueueConsumer {
	println("demoQueue QueueConsumer")
	return nil
}

func (d demoFilter) InitPlugin() error {
	println("demoFilter init")
	return nil
}

func (d demoFilter) Process(in api.Event) api.Event {
	println("demoFilter Process")
	return nil
}

func (d demoForwarder) InitPlugin() error {
	println("demoForwarder Process")
	return nil
}

func (d demoForwarder) Prepare() {
	println("demoForwarder Prepare")
}

func (d demoForwarder) Close() error {
	println("demoForwarder Close")
	return nil
}

func (d demoForwarder) Forward(batch api.BatchEvents) {
	println("demoForwarder Forward")
}

func (d demoForwarder) ForwardType() api.EventType {
	println("demoForwarder ForwardType")
	return api.SegmentEvent
}

func (d demoFallbacker) InitPlugin() error {
	println("demoFallbacker init")
	return nil
}

func (d demoFallbacker) FallBack(batch api.BatchEvents) api.Fallbacker {
	println("demoFallbacker FallBack")
	return nil
}

func (d demoClient) InitPlugin() error {
	println("demoClient init")
	return nil
}

func (d demoClient) Prepare() {
	println("demoClient Prepare")
}

func (d demoClient) Close() error {
	println("demoClient Close")
	return nil
}

func TestRegisterPlugin1(t *testing.T) {

}
