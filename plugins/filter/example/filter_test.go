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

package example

import (
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/filter/deinefilter"
)

func Test_Register(t *testing.T) {
	tests := []struct {
		name  string
		args  interface{}
		panic bool
	}{
		{
			name: "demoFilter",
			args: &demoFilter{
				a: "s",
			},
			panic: false,
		},
		{
			name: "demoFilter2",
			args: demoFilter2{
				a: "s",
			},
			panic: false,
		},
		{
			name: "demoFilter3",
			args: demoFilter3{
				a: "s",
			},
			panic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin.RegisterPlugin(tt.name, tt.args)
			assertPanic(t, tt.name, nil, tt.panic)
		})
	}
}

func assertPanic(t *testing.T, name string, config map[string]interface{}, existPanic bool) {
	defer func() {
		if r := recover(); r != nil && !existPanic {
			t.Errorf("the plugin %s is not pass", name)
		}
	}()
	deinefilter.GetFilter(name, config)
}
