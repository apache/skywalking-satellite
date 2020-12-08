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

type GathererConfig struct {
	// common config
	common ModuleCommonConfig

	// plugins config
	CollectorConfig plugin.DefaultConfig `mapstructure:"collector"` // collector plugin config
	QueueConfig     plugin.DefaultConfig `mapstructure:"queue"`     // queue plugin config
}

func (g *GathererConfig) ModuleName() string {
	return constant.GathererModule
}

func (g *GathererConfig) NameSpace() string {
	return g.common.RunningNamespace
}
