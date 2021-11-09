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

package none

import (
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

const (
	Name     = "none-fallbacker"
	ShowName = "None Fallbacker"
)

type Fallbacker struct {
	config.CommonFields
}

func (f *Fallbacker) Name() string {
	return Name
}

func (f *Fallbacker) ShowName() string {
	return ShowName
}

func (f *Fallbacker) Description() string {
	return "The fallbacker would do nothing when facing failure data."
}

func (f *Fallbacker) DefaultConfig() string {
	return ""
}

func (f *Fallbacker) FallBack(batch event.BatchEvents, forward api.ForwardFunc) bool {
	return true
}
