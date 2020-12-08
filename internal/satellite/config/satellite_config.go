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
)

// SatelliteConfig is to initialize Satellite.
type SatelliteConfig struct {
	Logger *log.LoggerConfig `mapstructure:"logger"`
	Agents []*AgentConfig    `mapstructure:"agents"`
}

// AgentConfig initializes the different module in different namespace.
type AgentConfig struct {
	ModuleCommonConfig *ModuleCommonConfig  `mapstructure:"common_config"`
	ClientManager      *ClientManagerConfig `mapstructure:"client_manager"`
	Gatherer           *GathererConfig      `mapstructure:"gatherer"`
	Processor          *ProcessorConfig     `mapstructure:"processor"`
	Sender             *SenderConfig        `mapstructure:"sender"`
}

// ModuleConfig is an interface to define the initialization fields in ModuleService.
type ModuleConfig interface {
	// ModuleName returns the name of current module.
	ModuleName() string
	// NameSpace returns the current running namespace. Satellite support multi-agents running in one Satellite
	// process, the namespace is a space concept to distinguish different agents.
	NameSpace() string
}

// ModuleCommonConfig has some common fields of each ModuleConfig.
type ModuleCommonConfig struct {
	RunningNamespace string `mapstructure:"namespace"`
}
