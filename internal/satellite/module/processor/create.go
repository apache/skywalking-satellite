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

package processor

import (
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/processor/api"
	sender "github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	filter "github.com/apache/skywalking-satellite/plugins/filter/api"
)

// Init Processor and dependency plugins
func NewProcessor(cfg *api.ProcessorConfig, s sender.Sender, g gatherer.Gatherer) api.Processor {
	log.Logger.Infof("processor module of %s namespace is being initialized", cfg.PipeName)
	p := &Processor{
		sender:         s,
		gatherer:       g,
		config:         cfg,
		runningFilters: []filter.Filter{},
	}
	for _, c := range p.config.FilterConfig {
		p.runningFilters = append(p.runningFilters, filter.GetFilter(c))
	}
	return p
}
