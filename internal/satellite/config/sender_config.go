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
	"github.com/apache/skywalking-satellite/internal/pkg/constant"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
)

type SenderConfig struct {
	// common config
	common ModuleCommonConfig

	// plugins config
	ForwardersConfig []plugin.DefaultConfig `mapstructure:"forwarders"` // forwarder plugins config
	FallbackerConfig plugin.DefaultConfig   `mapstructure:"fallbacker"` // fallbacker plugins config

	// self config
	MaxBufferSize  int `mapstructure:"max_buffer_size"`  // the max buffer capacity
	MinFlushEvents int `mapstructure:"min_flush_events"` // the min flush events when receives a timer flush signal
	FlushTime      int `mapstructure:"flush_time"`       // the period flush time
}

func (s *SenderConfig) ModuleName() string {
	return constant.SenderModule
}

func (s *SenderConfig) NameSpace() string {
	return s.common.RunningNamespace
}
