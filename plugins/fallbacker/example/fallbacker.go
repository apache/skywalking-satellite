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
	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/plugins/fallbacker/api"
)

type demoFallbacker struct {
	a string
}

type demoFallbacker2 struct {
	a string
}

type demoFallbacker3 struct {
	a string
}

func (d *demoFallbacker) Description() string {
	panic("implement me")
}

func (d *demoFallbacker) InitPlugin(config map[string]interface{}) {
}

func (d *demoFallbacker) FallBack(batch event.BatchEvents) api.Fallbacker {
	panic("implement me")
}

func (d demoFallbacker2) Description() string {
	panic("implement me")
}

func (d demoFallbacker2) InitPlugin(config map[string]interface{}) {
}

func (d demoFallbacker2) FallBack(batch event.BatchEvents) api.Fallbacker {
	panic("implement me")
}
