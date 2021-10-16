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
	"context"
	"reflect"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// Fetcher is a plugin interface, that defines new fetchers.
type Fetcher interface {
	plugin.Plugin

	Prepare()
	// Fetch would fetch some APM data.
	Fetch(ctx context.Context)
	// Channel would be put a data when the receiver receives an APM data.
	Channel() <-chan *v1.SniffData
	// Shutdown shutdowns the fetcher
	Shutdown(context.Context) error
	// SupportForwarders should provider all forwarder support current receiver
	SupportForwarders() []forwarder.Forwarder
}

// GetFetcher gets an initialized fetcher plugin.
func GetFetcher(config plugin.Config) Fetcher {
	return plugin.Get(reflect.TypeOf((*Fetcher)(nil)).Elem(), config).(Fetcher)
}
