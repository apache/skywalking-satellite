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

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	processor "github.com/apache/skywalking-satellite/internal/satellite/module/processor/api"
	sender "github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
)

func TestLoad(t *testing.T) {
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
			args: args{configPath: "../../../configs/satellite_config.yaml"},
			want: params(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := load(tt.args.configPath)
			if err != nil {
				t.Fatalf("cannot load config: %v", err)
			}
			doJudgeEqual(t, c.Logger, tt.want.Logger)
			doJudgeEqual(t, c.Sharing.Servers, tt.want.Sharing.Servers)
			doJudgeEqual(t, c.Sharing.Clients, tt.want.Sharing.Clients)
			doJudgeEqual(t, c.Pipes[0].PipeCommonConfig, tt.want.Pipes[0].PipeCommonConfig)
			doJudgeEqual(t, c.Pipes[0].Gatherer, tt.want.Pipes[0].Gatherer)
			doJudgeEqual(t, c.Pipes[0].Processor, tt.want.Pipes[0].Processor)
			doJudgeEqual(t, c.Pipes[0].Sender, tt.want.Pipes[0].Sender)
		})
	}
}

func params() *SatelliteConfig {
	return &SatelliteConfig{
		Logger: &log.LoggerConfig{
			LogPattern:  "%time [%level][%field] - %msg",
			TimePattern: "2006-01-02 15:04:05.000",
			Level:       "info",
		},
		Telemetry: &telemetry.Config{
			Cluster:  "cluster1",
			Service:  "service1",
			Instance: "instance1",
		},
		Sharing: sharing(),
		Pipes:   pipes(),
	}
}

func sharing() *SharingConfig {
	return &SharingConfig{
		SharingCommonConfig: &config.CommonFields{
			PipeName: "sharing",
		},
		Clients: []plugin.Config{
			{
				"plugin_name":            "kafka-client",
				"brokers":                "127.0.0.1:9092",
				"version":                "2.1.1",
				"commonfields_pipe_name": "sharing",
				"ca_pem_path":            "ca.pem",
				"client_key_path":        "client.key",
				"client_pem_path":        "client.pem",
				"enable_TLS":             false,
				"insecure_skip_verify":   false,
			},
			{
				"plugin_name":            "grpc-client",
				"server_addr":            "127.0.0.1:11800",
				"commonfields_pipe_name": "sharing",
				"ca_pem_path":            "ca.pem",
				"client_key_path":        "client.key",
				"client_pem_path":        "client.pem",
				"enable_TLS":             false,
				"insecure_skip_verify":   false,
				"check_period":           5,
				"authentication":         "",
			},
		},
		Servers: []plugin.Config{
			{
				"plugin_name":            "grpc-server",
				"commonfields_pipe_name": "sharing",
				"address":                ":11800",
				"tls_cert_file":          "",
				"tls_key_file":           "",
			},
			{
				"plugin_name":            "prometheus-server",
				"address":                ":1234",
				"commonfields_pipe_name": "sharing",
				"endpoint":               "/metrics",
			},
		},
	}
}

func pipes() []*PipeConfig {
	return []*PipeConfig{
		{
			PipeCommonConfig: &config.CommonFields{
				PipeName: "logpipe",
			},

			Gatherer: &gatherer.GathererConfig{
				ServerName: "grpc-server",
				CommonFields: &config.CommonFields{
					PipeName: "logpipe",
				},
				ReceiverConfig: plugin.Config{
					"plugin_name":            "grpc-nativelog-receiver",
					"commonfields_pipe_name": "logpipe",
				},
				QueueConfig: plugin.Config{
					"commonfields_pipe_name": "logpipe",
					"plugin_name":            "memory-queue",
					"event_buffer_size":      5000,
				},
			},
			Processor: &processor.ProcessorConfig{
				CommonFields: &config.CommonFields{
					PipeName: "logpipe",
				},
			},
			Sender: &sender.SenderConfig{
				CommonFields: &config.CommonFields{
					PipeName: "logpipe",
				},
				FallbackerConfig: plugin.Config{
					"commonfields_pipe_name": "logpipe",
					"plugin_name":            "none-fallbacker",
				},
				FlushTime:      1000,
				MaxBufferSize:  200,
				MinFlushEvents: 100,
				ClientName:     "kafka-client",
				ForwardersConfig: []plugin.Config{
					{
						"plugin_name":            "nativelog-kafka-forwarder",
						"topic":                  "log-topic",
						"commonfields_pipe_name": "logpipe",
					},
				},
			},
		},
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
		t.Fatalf("config is not equal, got %s,\n want %s", ajson, bjson)
	}
}
