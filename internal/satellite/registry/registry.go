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
	"fmt"

	"github.com/apache/skywalking-satellite/internal/pkg/api"
)

// The creator registry.
// All plugins is wrote in ./plugins dir. The plugin type would be as the next level dirs,
// such as collector, client, or queue. And the 3rd level is the plugin name, that is also
// used as key in pluginRegistry.
type pluginRegistry struct {
	collectorCreatorRegistry map[string]CollectorCreator
	queueCreatorRegistry     map[string]QueueCreator
	filterCreatorRegistry    map[string]FilterCreator
	forwarderCreatorRegistry map[string]ForwarderCreator
	parserCreatorRegistry    map[string]ParserCreator
	clientCreatorRegistry    map[string]ClientCreator
}

// ClientCreator creates a Client according to the config.
type ClientCreator func(config map[string]interface{}) (api.Client, error)

// CollectorCreator creates a Collector according to the config.
type CollectorCreator func(config map[string]interface{}) (api.Collector, error)

// QueueCreator creates a Queue according to the config.
type QueueCreator func(config map[string]interface{}) (api.Queue, error)

// FilterCreator creates a Filter according to the config.
type FilterCreator func(config map[string]interface{}) (api.Filter, error)

// ForwarderCreator creates a forwarder according to the config.
type ForwarderCreator func(config map[string]interface{}) (api.Forwarder, error)

// ParserCreator creates a parser according to the config.
type ParserCreator func(config map[string]interface{}) (api.Parser, error)

var reg *pluginRegistry

// RegisterClient registers the clientType as ClientCreator.
func RegisterClient(clientType string, creator ClientCreator) {
	fmt.Printf("Create %s client creator register successfully", clientType)
	reg.clientCreatorRegistry[clientType] = creator
}

// RegisterCollector registers the collectorType as CollectorCreator.
func RegisterCollector(collectorType string, creator CollectorCreator) {
	fmt.Printf("Create %s collector creator register successfully", collectorType)
	reg.collectorCreatorRegistry[collectorType] = creator
}

// RegisterQueue registers the queueType as QueueCreator.
func RegisterQueue(queueType string, creator QueueCreator) {
	fmt.Printf("Create %s queue creator register successfully", queueType)
	reg.queueCreatorRegistry[queueType] = creator
}

// RegisterFilter registers the filterType as FilterCreator.
func RegisterFilter(filterType string, creator FilterCreator) {
	fmt.Printf("Create %s filter creator register successfully", filterType)
	reg.filterCreatorRegistry[filterType] = creator
}

// RegisterForwarder registers the forwarderType as forwarderCreator.
func RegisterForwarder(forwarderType string, creator ForwarderCreator) {
	fmt.Printf("Create %s forward creator register successfully", forwarderType)
	reg.forwarderCreatorRegistry[forwarderType] = creator
}

// CreateClient creates a Client according to the clientType.
func CreateClient(clientType string, config map[string]interface{}) (api.Client, error) {
	if c, ok := reg.clientCreatorRegistry[clientType]; ok {
		client, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create client failed: %v", err)
		}
		return client, nil
	}
	return nil, fmt.Errorf("unsupported client type: %v", clientType)
}

// CreateCollector creates a Collector according to the collectorType.
func CreateCollector(collectorType string, config map[string]interface{}) (api.Collector, error) {
	if c, ok := reg.collectorCreatorRegistry[collectorType]; ok {
		collector, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create collector failed: %v", err)
		}
		return collector, nil
	}
	return nil, fmt.Errorf("unsupported collector type: %v", collectorType)
}

// CreateQueue creates a Queue according to the queueType.
func CreateQueue(queueType string, config map[string]interface{}) (api.Queue, error) {
	if c, ok := reg.queueCreatorRegistry[queueType]; ok {
		queue, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create queue failed: %v", err)
		}
		return queue, nil
	}
	return nil, fmt.Errorf("unsupported queue type: %v", queueType)
}

// CreateFilter creates a Filter according to the filterType.
func CreateFilter(filterType string, config map[string]interface{}) (api.Filter, error) {
	if c, ok := reg.filterCreatorRegistry[filterType]; ok {
		filter, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create filter failed: %v", err)
		}
		return filter, nil
	}
	return nil, fmt.Errorf("unsupported filter type: %v", filterType)
}

// CreateForwarder creates a forwarder according to the forwarderType.
func CreateForwarder(forwarderType string, config map[string]interface{}) (api.Forwarder, error) {
	if c, ok := reg.forwarderCreatorRegistry[forwarderType]; ok {
		forwarder, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create forwarder failed: %v", err)
		}
		return forwarder, nil
	}
	return nil, fmt.Errorf("unsupported forwarder type: %v", forwarderType)
}

// CreateParser creates a parser according to the parserType.
func CreateParser(parserType string, config map[string]interface{}) (api.Parser, error) {
	if c, ok := reg.parserCreatorRegistry[parserType]; ok {
		parser, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create parser failed: %v", err)
		}
		return parser, nil
	}
	return nil, fmt.Errorf("unsupported parser type: %v", parserType)
}

func init() {
	if reg == nil {
		reg = &pluginRegistry{}
		reg.collectorCreatorRegistry = make(map[string]CollectorCreator)
		reg.queueCreatorRegistry = make(map[string]QueueCreator)
		reg.filterCreatorRegistry = make(map[string]FilterCreator)
		reg.forwarderCreatorRegistry = make(map[string]ForwarderCreator)
		reg.parserCreatorRegistry = make(map[string]ParserCreator)
		reg.clientCreatorRegistry = make(map[string]ClientCreator)
	}
}
