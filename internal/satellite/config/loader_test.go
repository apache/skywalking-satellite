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

package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/pkg/test"
)

func TestLoad(t *testing.T) {
	rootPath, err := test.FindRootPath()
	if err != nil {
		panic(err)
	}
	type args struct {
		configPath string
	}
	tests := []struct {
		name string
		args args
		want *SatelliteConfig
	}{
		{
			name: "Legal configuration",
			args: args{configPath: rootPath + "/configs/satellite_config.yaml"},
			want: &SatelliteConfig{
				Logger: &log.LoggerConfig{
					LogPattern:  "%time [%level][%field] - %msg",
					TimePattern: "2006-01-02 15:04:05.001",
					Level:       "info",
				},
				Agents: []*AgentConfig{
					{
						ModuleCommonConfig: &ModuleCommonConfig{
							RunningNamespace: "namespace1",
						},

						Gatherer: &GathererConfig{
							CollectorConfig: plugin.DefaultConfig{
								"plugin_name": "segment-receiver",
								"k":           "v",
							},
							QueueConfig: plugin.DefaultConfig{
								"plugin_name": "mmap-queue",
								"key":         "value",
							},
						},
						Processor: &ProcessorConfig{
							FilterConfig: []plugin.DefaultConfig{
								{
									"plugin_name": "filtertype1",
									"key":         "value",
								},
							},
						},
						Sender: &SenderConfig{
							FlushTime:     5000,
							MaxBufferSize: 100,
							ForwardersConfig: []plugin.DefaultConfig{
								{
									"plugin_name": "segment-forwarder",
									"key":         "value",
								},
							},
						},
					},
					{
						ModuleCommonConfig: &ModuleCommonConfig{
							RunningNamespace: "sharing",
						},
						ClientManager: &ClientManagerConfig{
							RetryInterval: 5000,
							ClientConfig: plugin.DefaultConfig{
								"plugin_name": "grpc",
								"k":           "v",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := load(tt.args.configPath)
			if err != nil {
				t.Fatalf("cannot load config: %v", err)
			}
			doJudgeEqual(t, c.Logger, tt.want.Logger)
			doJudgeEqual(t, c.Agents[0].ModuleCommonConfig, tt.want.Agents[0].ModuleCommonConfig)
			doJudgeEqual(t, c.Agents[0].Gatherer, tt.want.Agents[0].Gatherer)
			doJudgeEqual(t, c.Agents[0].Processor, tt.want.Agents[0].Processor)
			doJudgeEqual(t, c.Agents[0].Sender, tt.want.Agents[0].Sender)

			doJudgeEqual(t, c.Agents[1].ModuleCommonConfig, tt.want.Agents[1].ModuleCommonConfig)
			doJudgeEqual(t, c.Agents[1].ClientManager, tt.want.Agents[1].ClientManager)
		})
	}
}

func doJudgeEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		ajson, err := json.Marshal(a)
		if err != nil {
			t.Fatalf("cannot do json format: %v", err)
		}
		bjson, err := json.Marshal(b)
		if err != nil {
			t.Fatalf("cannot do json format: %v", err)
		}
		t.Fatalf("config is not equal, got %s, want %s", ajson, bjson)
	}
}
