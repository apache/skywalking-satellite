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

package resolvers

import (
	"strconv"

	"github.com/prometheus/prometheus/discovery/kubernetes"

	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/prometheus/common/model"
)

type PodAnalyzer struct {
}

func (p *PodAnalyzer) KindType() string {
	return string(kubernetes.RolePod)
}

func (p *PodAnalyzer) GetAddresses(cache map[string]*targetgroup.Group, config *KubernetesConfig) []string {
	result := make([]string, 0)
	for _, group := range cache {
		for _, target := range group.Targets {
			val, exists := target[model.LabelName("__meta_kubernetes_pod_container_port_number")]
			if exists && string(val) == strconv.Itoa(config.ExtraPort.Port) {
				result = append(result, string(target[model.LabelName("__address__")]))
			}
		}
	}
	return result
}
