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
	"context"
	"fmt"
	"net/url"

	"github.com/prometheus/prometheus/discovery"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	"google.golang.org/grpc/resolver"

	"github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

var analyzers = []KindAddressAnalyzer{
	&PodAnalyzer{},
	&ServiceAnalyzer{},
	&EndpointsAnalyzer{},
}

type KindCache struct {
	config   *KubernetesConfig
	cache    map[string]*targetgroup.Group
	cc       resolver.ClientConn
	analyzer KindAddressAnalyzer
}

type KindAddressAnalyzer interface {
	KindType() string
	GetAddresses(cache map[string]*targetgroup.Group, config *KubernetesConfig) []string
}

func NewKindCache(ctx context.Context, c *KubernetesConfig, cc resolver.ClientConn) (*KindCache, error) {
	// build config
	conf := &kubernetes.SDConfig{}
	if c.APIServer != "" {
		parsed, err := url.Parse(c.APIServer)
		if err != nil {
			return nil, err
		}
		conf.APIServer = config.URL{URL: parsed}
		httpConfig, err := c.HTTPClientConfig.convertHTTPConfig()
		if err != nil {
			return nil, err
		}
		conf.HTTPClientConfig = *httpConfig
	}

	conf.Role = kubernetes.Role(c.Kind)
	conf.NamespaceDiscovery = kubernetes.NamespaceDiscovery{
		Names: c.Namespaces,
	}
	conf.Selectors = []kubernetes.SelectorConfig{
		{
			Role:  kubernetes.Role(c.Kind),
			Label: c.Selector.Label,
			Field: c.Selector.Field,
		},
	}

	// build discovery
	discoverer, err := kubernetes.New(&logAdapt{}, conf)
	if err != nil {
		return nil, err
	}

	// build analyzer
	var analyzer KindAddressAnalyzer
	for _, a := range analyzers {
		if a.KindType() == c.Kind {
			analyzer = a
		}
	}
	if analyzer == nil {
		return nil, fmt.Errorf("could not kind analyzer: %s", c.Kind)
	}

	// build and start watch
	kind := &KindCache{config: c, cc: cc, analyzer: analyzer}
	kind.watchAndUpdate(ctx, discoverer)
	return kind, nil
}

func (w *KindCache) watchAndUpdate(ctx context.Context, discoverer discovery.Discoverer) {
	ch := make(chan []*targetgroup.Group)
	go discoverer.Run(ctx, ch)

	w.cache = make(map[string]*targetgroup.Group)
	go func() {
		for {
			select {
			case tgs := <-ch:
				for _, tg := range tgs {
					if tg.Targets == nil || len(tg.Targets) == 0 {
						delete(w.cache, tg.Source)
						continue
					}
					w.cache[tg.Source] = tg

					// dynamic update addresses
					if err := w.UpdateAddresses(); err != nil {
						log.Logger.Warnf("dynamic update addresss failure, %v", err)
					}
				}
			case <-ctx.Done():
				break
			}
		}
	}()
}

func (w *KindCache) UpdateAddresses() error {
	addresses := w.analyzer.GetAddresses(w.cache, w.config)
	addrs := make([]resolver.Address, len(addresses))
	for i, s := range addresses {
		addrs[i] = resolver.Address{Addr: s}
	}
	if err := w.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		return err
	}
	log.Logger.Infof("update grpc client addresses: %v", addresses)
	return nil
}

type logAdapt struct {
}

func (l *logAdapt) Log(keyvals ...interface{}) error {
	log.Logger.Print(keyvals...)
	return nil
}
