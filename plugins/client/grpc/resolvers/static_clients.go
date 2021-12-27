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
	"fmt"
	"strings"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	"google.golang.org/grpc/resolver"
)

var staticServerSchema = "static"

type staticServerResolver struct {
}

func (s *staticServerResolver) Type() string {
	return staticServerSchema
}

func (s *staticServerResolver) BuildTarget(c *ServerFinderConfig) (string, error) {
	// build target using uri endpoint
	return fmt.Sprintf("%s:///%s", staticServerSchema, c.ServerAddr), nil
}

//nolint:gocritic // Implement for resolver.Target
func (*staticServerResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &staticResolver{
		target: target,
		cc:     cc,
	}
	r.analyzeClients()
	return r, nil
}

func (*staticServerResolver) Scheme() string {
	return staticServerSchema
}

type staticResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *staticResolver) ResolveNow(o resolver.ResolveNowOptions) {
	r.analyzeClients()
}

func (*staticResolver) Close() {
}

func (r *staticResolver) analyzeClients() {
	addresses := strings.Split(strings.TrimLeft(r.target.URL.Path, "/"), ",")
	addrs := make([]resolver.Address, len(addresses))
	for i, s := range addresses {
		addrs[i] = resolver.Address{Addr: s}
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		log.Logger.Warnf("error update static grpc client list: %v", err)
	}
}
