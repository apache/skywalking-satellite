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

import "github.com/apache/skywalking-satellite/internal/pkg/event"

type demoFilter struct {
	a string
}

type demoFilter2 struct {
	a string
}

type demoFilter3 struct {
	a string
}

func (d *demoFilter) Description() string {
	panic("implement me")
}

func (d *demoFilter) InitPlugin(config map[string]interface{}) {
}

func (d *demoFilter) Process(in event.Event) event.Event {
	panic("implement me")
}

func (d demoFilter2) Description() string {
	panic("implement me")
}

func (d demoFilter2) InitPlugin(config map[string]interface{}) {
}

func (d demoFilter2) Process(in event.Event) event.Event {
	panic("implement me")
}
