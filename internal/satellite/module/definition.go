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

package module

import (
	"context"
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
)

const (
	// If a module is stored in the sharing namespace, the modules would be shared with all namespaces in moduleContainer.
	SharingNamespaceName = "sharing"
)

// TODO add metrics func
// Service id a custom plugin interface, which defines the processing.
type Service interface {
	plugin.Plugin

	// Config returns the using ModuleConfig.
	Config() config.ModuleConfig

	// Init initialize the Module and register the instance to the registry.
	// In this stage, no components is running.
	Init(config config.ModuleConfig)

	// Prepare would inject the dependency module and do dependency initialization.
	Prepare() error

	// Boot would start the module and return error when started failed. When a stop signal received
	// or an exception occurs, the shutdown function would be called.
	Boot(ctx context.Context)

	// Shutdown could do some clean job to close Service.
	Shutdown()
}

var moduleContainer map[string]Service

// NewModuleService returns a new initialized Service and register it to the module container.
func NewModuleService(cfg config.ModuleConfig) {
	_ = plugin.Get(reflect.TypeOf((*Service)(nil)).Elem(), cfg).(Service)
}

// GetRunningModule returns a running Service.
func GetRunningModule(namespace, moduleName string) Service {
	if moduleService, ok := moduleContainer[namespace+moduleName]; ok {
		return moduleService
	}
	if moduleService, ok := moduleContainer[SharingNamespaceName+moduleName]; ok {
		return moduleService
	}
	return nil
}

func GetModuleContainer() map[string]Service {
	return moduleContainer
}

// Register the Service category to the plugin registry.
func init() {
	moduleContainer = make(map[string]Service)
	plugin.RegisterPluginCategory(reflect.TypeOf((*Service)(nil)).Elem(),
		func(cfg interface{}) string {
			// Get plugin name to find the specific Service plugin.
			return cfg.(config.ModuleConfig).ModuleName()
		},
		func(plugin plugin.Plugin, cfg interface{}) {
			// Initialize Service.
			ms := plugin.(Service)
			mc := cfg.(config.ModuleConfig)
			ms.Init(mc)
		},
		func(plugin plugin.Plugin) {
			// Register the initialized Service to the moduleContainer.
			ms := plugin.(Service)
			moduleContainer[ms.Config().NameSpace()+ms.Config().ModuleName()] = ms
		},
	)
}
