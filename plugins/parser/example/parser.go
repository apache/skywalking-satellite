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

type demoParser struct {
	a string
}

type demoParser2 struct {
	a string
}

type demoParser3 struct {
	a string
}

func (d *demoParser) Description() string {
	panic("implement me")
}

func (d *demoParser) InitPlugin(config map[string]interface{}) {
}

func (d *demoParser) ParseBytes(bytes []byte) ([]event.SerializableEvent, error) {
	panic("implement me")
}

func (d *demoParser) ParseStr(str string) ([]event.SerializableEvent, error) {
	panic("implement me")
}

func (d demoParser2) Description() string {
	panic("implement me")
}

func (d demoParser2) InitPlugin(config map[string]interface{}) {
}

func (d demoParser2) ParseBytes(bytes []byte) ([]event.SerializableEvent, error) {
	panic("implement me")
}

func (d demoParser2) ParseStr(str string) ([]event.SerializableEvent, error) {
	panic("implement me")
}
