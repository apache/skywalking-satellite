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

package api

import (
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
)

//   Init()     Initial stage: Init plugin by config
//    ||
//    \/
//   Prepare()   Preparing stage: Prepare the collector, such as build connection with SkyWalking javaagent.
//    ||
//    \/
//   Next()     Running stage: When Collector collect a data, the data would be fetched by the upstream
//    ||                       component through this method.
//    \/
//   Close()    Closing stage: Close the Collector, such as close connection with SkyWalking javaagent.

// Collector is a plugin interface, that defines new collectors.
type Collector interface {
	plugin.Plugin

	// Prepare creates a listen or reader to gather data.
	Prepare()
	// Next return the data from the input.
	Next() (event.SerializableEvent, error)
	// Close would close collector.
	Close()
}

var CollectorCategory = reflect.TypeOf((*Collector)(nil)).Elem()

// Get collector plugin.
func GetCollector(pluginName string, config map[string]interface{}) Collector {
	return plugin.Get(CollectorCategory, pluginName, config).(Collector)
}

func init() {
	plugin.AddPluginCategory(CollectorCategory)
}
