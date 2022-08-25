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

package otlpmetricsv1

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/client/grpc/lb"

	"k8s.io/apimachinery/pkg/util/cache"
	metrics "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/collector/metrics/v1"
	common "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/common/v1"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"google.golang.org/grpc"
)

const (
	Name     = "otlp-metrics-v1-grpc-forwarder"
	ShowName = "OpenTelemetry Metrics v1 GRPC Forwarder"
)

type Forwarder struct {
	config.CommonFields
	// The label key of the routing data, multiple keys are split by ","
	RoutingLabelKeys string `mapstructure:"routing_label_keys"`
	// The LRU policy cache size for hosting routine rules of service instance.
	RoutingRuleLRUCacheSize int `mapstructure:"routing_rule_lru_cache_size"`
	// The TTL of the LRU cache size for hosting routine rules of service instance.
	RoutingRuleLRUCacheTTL int `mapstructure:"routing_rule_lru_cache_ttl"`

	metricsClient       metrics.MetricsServiceClient
	metadataKeys        []string
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
	return "This is a synchronization grpc forwarder with the OpenTelemetry metrics v1 protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return `
# The LRU policy cache size for hosting routine rules of service instance.
routing_rule_lru_cache_size: 5000
# The TTL of the LRU cache size for hosting routine rules of service instance.
routing_rule_lru_cache_ttl: 180
# The label key of the routing data, multiple keys are split by ","
routing_label_keys: net.host.name,host.name,job,service.name
`
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.metricsClient = metrics.NewMetricsServiceClient(client)
	if f.RoutingLabelKeys == "" {
		return fmt.Errorf("please provide metadata keys")
	}
	f.metadataKeys = strings.Split(f.RoutingLabelKeys, ",")
	f.upstreamCache = cache.NewLRUExpireCache(f.RoutingRuleLRUCacheSize)
	f.upstreamCacheExpire = time.Second * time.Duration(f.RoutingRuleLRUCacheTTL)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	for _, d := range batch {
		key, err := f.generateRoutingKey(d.GetOpenTelementryMetricsV1Request())
		if err != nil {
			log.Logger.Errorf("generate the routing key failure: %v", err)
			continue
		}
		ctx := lb.WithLoadBalanceConfig(
			context.Background(),
			key,
			f.loadCachedPeer(key))

		_, err = f.metricsClient.Export(ctx, d.GetOpenTelementryMetricsV1Request())
		if err != nil {
			log.Logger.Errorf("%s send meter data error: %v", f.Name(), err)
			return err
		}
		f.savePeerInstanceFromStream(ctx, key)
	}
	return nil
}

func (f *Forwarder) savePeerInstanceFromStream(ctx context.Context, instance string) {
	upstream := lb.GetAddress(ctx)
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

func (f *Forwarder) generateRoutingKey(data *metrics.ExportMetricsServiceRequest) (string, error) {
	if len(data.GetResourceMetrics()) == 0 {
		return "", fmt.Errorf("no resources")
	}
	var lastKVs []*common.KeyValue
	for _, m := range data.GetResourceMetrics() {
		if m.Resource == nil {
			continue
		}
		if len(m.Resource.Attributes) == 0 {
			continue
		}
		lastKVs = m.Resource.Attributes
		result := ""
		for _, kv := range m.Resource.Attributes {
			for _, key := range f.metadataKeys {
				if kv.GetKey() == key {
					result += fmt.Sprintf(",%s", kv.GetValue().GetStringValue())
				}
			}
		}
		if result != "" {
			return result, nil
		}
	}
	if lastKVs == nil {
		return "", fmt.Errorf("could not found any attributes")
	}

	var keys string
	for i, k := range lastKVs {
		if i > 0 {
			keys += ","
		}
		keys += k.GetKey()
	}
	return "", fmt.Errorf("could not found anly routing key, existing keys sample: %s", keys)
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_OpenTelementryMetricsV1Type
}

func (f *Forwarder) SyncForward(*v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}
