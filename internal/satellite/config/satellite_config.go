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
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/module"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
)

// SatelliteConfig is to initialize Satellite.
type SatelliteConfig struct {
	Logger        *log.LoggerConfig           `mapstructure:"logger"`
	Agents        []*AgentConfig              `mapstructure:"agents"`
	ClientManager *module.ClientManagerConfig `mapstructure:"client_manager"`
}

// AgentConfig initializes the different module in different namespace.
type AgentConfig struct {
	ModuleCommonConfig *api.ModuleCommonConfig     `mapstructure:"common_config"`
	ClientManager      *module.ClientManagerConfig `mapstructure:"client_manager"`
	Gatherer           *module.GathererConfig      `mapstructure:"gatherer"`
	Processor          *module.ProcessorConfig     `mapstructure:"processor"`
	Sender             *module.SenderConfig        `mapstructure:"sender"`
}
