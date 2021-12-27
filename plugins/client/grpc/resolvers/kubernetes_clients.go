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
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

var kubernetesServerSchema = "kubernetes"

type kubernetesServerResolver struct {
}

func (k *kubernetesServerResolver) Type() string {
	return kubernetesServerSchema
}

func (k *kubernetesServerResolver) BuildTarget(c *ServerFinderConfig) (string, error) {
	marshal, err := json.Marshal(c.KubernetesConfig)
	if err != nil {
		return "", fmt.Errorf("convert kubernetes config error: %v", err)
	}
	return fmt.Sprintf("%s:///%s", kubernetesServerSchema, string(marshal)), nil
}

//nolint:gocritic // Implement for resolver.Target
func (*kubernetesServerResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// convert data
	kubernetesConfig := &KubernetesConfig{}
	if err := json.Unmarshal([]byte(strings.TrimLeft(target.URL.Path, "/")), kubernetesConfig); err != nil {
		return nil, fmt.Errorf("could not analyze the address: %v", err)
	}

	// validate http config
	if kubernetesConfig.APIServer != "" {
		httpConfig, err := kubernetesConfig.HTTPClientConfig.convertHTTPConfig()
		if err != nil {
			return nil, err
		}
		if err = httpConfig.Validate(); err != nil {
			return nil, fmt.Errorf("http config validate error: %v", err)
		}
	}

	// init cache
	ctx, cancel := context.WithCancel(context.Background())
	cache, err := NewKindCache(ctx, kubernetesConfig, cc)
	if err != nil {
		cancel()
		return nil, err
	}

	// build resolver
	r := &kubernetesResolver{
		cache:  cache,
		cancel: cancel,
	}
	return r, nil
}

func (*kubernetesServerResolver) Scheme() string {
	return kubernetesServerSchema
}

type kubernetesResolver struct {
	cache  *KindCache
	cancel context.CancelFunc
}

func (k *kubernetesResolver) ResolveNow(o resolver.ResolveNowOptions) {
	if err := k.cache.UpdateAddresses(); err != nil {
		log.Logger.Warnf("error update static grpc client list: %v", err)
	}
}

func (k *kubernetesResolver) Close() {
	k.cancel()
}
