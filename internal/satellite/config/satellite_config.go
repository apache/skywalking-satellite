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
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	processor "github.com/apache/skywalking-satellite/internal/satellite/module/processor/api"
	sender "github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
)

// SatelliteConfig is to initialize Satellite.
type SatelliteConfig struct {
	Logger    *log.LoggerConfig `mapstructure:"logger"`
	Pipes     []*PipeConfig     `mapstructure:"pipes"`
	Sharing   *SharingConfig    `mapstructure:"sharing"`
	Telemetry *telemetry.Config `mapstructure:"telemetry"`
}

// SharingConfig contains some plugins,which could be shared by every namespace. That is useful to reduce resources cost.
type SharingConfig struct {
	SharingCommonConfig config.CommonFields `mapstructure:"common_config"`
	Clients             []plugin.Config     `mapstructure:"clients"`
	Servers             []plugin.Config     `mapstructure:"servers"`
}

// PipeConfig initializes the different module in different namespace.
type PipeConfig struct {
	PipeCommonConfig config.CommonFields        `mapstructure:"common_config"`
	Gatherer         *gatherer.GathererConfig   `mapstructure:"gatherer"`
	Processor        *processor.ProcessorConfig `mapstructure:"processor"`
	Sender           *sender.SenderConfig       `mapstructure:"sender"`
}
