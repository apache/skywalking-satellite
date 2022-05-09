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

package nativemeter

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/cache"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/client/grpc/lb"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"
)

const (
	Name     = "native-meter-grpc-forwarder"
	ShowName = "Native Meter GRPC Forwarder"
)

type Forwarder struct {
	config.CommonFields
	// The LRU policy cache size for hosting routine rules of service instance.
	RoutingRuleLRUCacheSize int `mapstructure:"routing_rule_lru_cache_size"`
	// The TTL of the LRU cache size for hosting routine rules of service instance.
	RoutingRuleLRUCacheTTL int `mapstructure:"routing_rule_lru_cache_ttl"`

	meterClient         v3.MeterReportServiceClient
	upstreamCache       *cache.LRUExpireCache
	upstreamCacheExpire time.Duration
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a synchronization meter grpc forwarder with the SkyWalking meter protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return `
# The LRU policy cache size for hosting routine rules of service instance.
routing_rule_lru_cache_size: 5000
# The TTL of the LRU cache size for hosting routine rules of service instance.
routing_rule_lru_cache_ttl: 180
`
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.meterClient = v3.NewMeterReportServiceClient(client)
	f.upstreamCache = cache.NewLRUExpireCache(f.RoutingRuleLRUCacheSize)
	f.upstreamCacheExpire = time.Second * time.Duration(f.RoutingRuleLRUCacheTTL)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	streamMap := make(map[string]grpc.ClientStream)
	defer func() {
		for _, stream := range streamMap {
			err := closeStream(stream)
			if err != nil {
				log.Logger.Warnf("%s close stream error: %v", f.Name(), err)
			}
		}
	}()
	for _, e := range batch {
		// Only handle the meter collection data from queue
		// There could have error when using previously meter data(SniffData_Meter)
		if data, ok := e.GetData().(*v1.SniffData_MeterCollection); ok {
			if err := f.handleMeterCollection(data, streamMap); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Forwarder) handleMeterCollection(data *v1.SniffData_MeterCollection, streamMap map[string]grpc.ClientStream) error {
	if len(data.MeterCollection.MeterData) == 0 {
		return nil
	}
	firstMeter := data.MeterCollection.MeterData[0]
	streamName := fmt.Sprintf("batch-stream-%s-%s", firstMeter.Service, firstMeter.ServiceInstance)
	stream := streamMap[streamName]
	if stream == nil {
		ctx := lb.WithLoadBalanceConfig(
			context.Background(),
			firstMeter.ServiceInstance,
			f.loadCachedPeer(firstMeter.ServiceInstance))

		curStream, err := f.meterClient.CollectBatch(ctx)
		if err != nil {
			log.Logger.Errorf("open grpc stream error %v", err)
			return err
		}
		streamMap[streamName] = curStream
		stream = curStream
		f.savePeerInstanceFromStream(curStream, firstMeter.ServiceInstance)
	}

	if err := stream.SendMsg(data.MeterCollection); err != nil {
		log.Logger.Errorf("%s send meter data error: %v", f.Name(), err)
		return err
	}
	return nil
}

func (f *Forwarder) savePeerInstanceFromStream(stream grpc.ClientStream, instance string) {
	upstream := server_grpc.GetPeerAddressFromStreamContext(stream.Context())
	if upstream == "" {
		return
	}

	f.upstreamCache.Add(instance, upstream, f.upstreamCacheExpire)
}

func (f *Forwarder) loadCachedPeer(instance string) string {
	if get, exists := f.upstreamCache.Get(instance); exists {
		return get.(string)
	}
	return ""
}

func closeStream(stream grpc.ClientStream) error {
	if err := stream.CloseSend(); err != nil && err != io.EOF {
		return err
	}
	if err := stream.RecvMsg(server_grpc.NewOriginalData(nil)); err != nil {
		return err
	}
	return nil
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_MeterType
}

func (f *Forwarder) SyncForward(_ *v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}
