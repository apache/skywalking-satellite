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

package lb

import (
	"hash/crc32"
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Name = "satellite_lb"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &satelliteDynamicPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type satelliteDynamicPickerBuilder struct {
}

func (s *satelliteDynamicPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	addrToConn := make(map[string]balancer.SubConn)
	cons := make([]*connectionWrap, 0)
	for conn, connInfo := range info.ReadySCs {
		addrToConn[connInfo.Address.Addr] = conn
		cons = append(cons, &connectionWrap{
			addr: connInfo.Address.Addr,
			conn: conn,
		})
	}

	return &satelliteDynamicPicker{
		cons:       cons,
		addrToConn: addrToConn,
		connCount:  len(cons),
	}
}

type satelliteDynamicPicker struct {
	cons       []*connectionWrap
	addrToConn map[string]balancer.SubConn
	connCount  int

	mu   sync.Mutex
	next int
}

type connectionWrap struct {
	addr string
	conn balancer.SubConn
}

func (s *satelliteDynamicPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// only one connection
	if s.connCount == 1 {
		log.Logger.Debugf("pick the connection: %s", s.cons[0].addr)
		return balancer.PickResult{SubConn: s.cons[0].conn}, nil
	}

	config := queryConfig(info.Ctx)
	// if not exists config just round-robin the connection
	if config == nil {
		return s.roundRobinConnection(), nil
	}

	// check exists appoint address
	if config.appointAddr != "" {
		if con := s.addrToConn[config.appointAddr]; con != nil {
			log.Logger.Debugf("use the appoint connection: %s", config.appointAddr)
			return balancer.PickResult{SubConn: con}, nil
		}
	}

	// hash the route key
	routeIndex := hashCode(config.routeKey) % s.connCount
	connWrap := s.cons[routeIndex]
	// update the address to the config
	config.appointAddr = connWrap.addr
	log.Logger.Debugf("pick the connection: %s", connWrap.addr)
	return balancer.PickResult{SubConn: connWrap.conn}, nil
}

func (s *satelliteDynamicPicker) roundRobinConnection() balancer.PickResult {
	s.mu.Lock()
	sc := s.cons[s.next]
	s.next = (s.next + 1) % s.connCount
	s.mu.Unlock()
	log.Logger.Debugf("pick the connection: %s", sc.addr)
	return balancer.PickResult{SubConn: sc.conn}
}

func hashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}
