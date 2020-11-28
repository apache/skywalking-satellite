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
	"github.com/apache/skywalking-satellite/internal/pkg/logger"
)

// The creator reg.
type registry struct {
	gathererCreatorRegistry  map[string]GathererCreator
	queueCreatorRegistry     map[string]QueueCreator
	filterCreatorRegistry    map[string]FilterCreator
	forwarderCreatorRegistry map[string]ForwarderCreator
	parserCreatorRegistry    map[string]ParserCreator
}

// GathererCreator creates a Gather according to the config.
type GathererCreator func(config map[string]interface{}) (api.Gatherer, error)

// QueueCreator creates a Queue according to the config.
type QueueCreator func(config map[string]interface{}) (api.Queue, error)

// FilterCreator creates a Filter according to the config.
type FilterCreator func(config map[string]interface{}) (api.Filter, error)

// ForwarderCreator creates a forwarder according to the config.
type ForwarderCreator func(config map[string]interface{}) (api.Forwarder, error)

// ParserCreator creates a parser according to the config.
type ParserCreator func(config map[string]interface{}) (api.Parser, error)

var reg *registry

// RegisterGatherer registers the gatherType as GathererCreator.
func RegisterGatherer(gathererType string, creator GathererCreator) {
	logger.Log.Info(gathererType)
	reg.gathererCreatorRegistry[gathererType] = creator
}

// RegisterQueue registers the queueType as QueueCreator.
func RegisterQueue(queueType string, creator QueueCreator) {
	reg.queueCreatorRegistry[queueType] = creator
}

// RegisterFilter registers the filterType as FilterCreator.
func RegisterFilter(filterType string, creator FilterCreator) {
	reg.filterCreatorRegistry[filterType] = creator
}

// RegisterForwarder registers the forwarderType as forwarderCreator.
func RegisterForwarder(forwarderType string, creator ForwarderCreator) {
	reg.forwarderCreatorRegistry[forwarderType] = creator
}

// CreateGatherer creates a Gatherer according to the gathererType.
func CreateGatherer(gathererType string, config map[string]interface{}) (api.Gatherer, error) {
	if c, ok := reg.gathererCreatorRegistry[gathererType]; ok {
		gatherer, err := c(config)
		if err != nil {
			return nil, fmt.Errorf("create gatherer failed: %v", err)
		}
		return gatherer, nil
	}
	return nil, fmt.Errorf("unsupported gatherer type: %v", gathererType)
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
		reg = &registry{}
		reg.gathererCreatorRegistry = make(map[string]GathererCreator)
		reg.queueCreatorRegistry = make(map[string]QueueCreator)
		reg.filterCreatorRegistry = make(map[string]FilterCreator)
		reg.forwarderCreatorRegistry = make(map[string]ForwarderCreator)
		reg.parserCreatorRegistry = make(map[string]ParserCreator)
	}
}
