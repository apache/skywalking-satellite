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

	"google.golang.org/grpc/resolver"
)

// all customized resolvers
var rs = []GrpcResolver{
	&staticServerResolver{},
	&kubernetesServerResolver{},
}

type ServerFinderConfig struct {
	FinderType string `mapstructure:"finder_type"` // The gRPC server address finder type, support "static" and "kubernetes"
	// The gRPC server address, only works for "static" address finder
	ServerAddr string `mapstructure:"server_addr"`
	// The kubernetes config to lookup addresses, only works for "kubernetes" address finder
	KubernetesConfig *KubernetesConfig `mapstructure:"kubernetes_config"`
}

type GrpcResolver interface {
	resolver.Builder

	// Type of resolver
	Type() string
	// BuildTarget address by client config
	BuildTarget(c *ServerFinderConfig) (string, error)
}

func RegisterAllGrpcResolvers() {
	for _, r := range rs {
		resolver.Register(r)
	}
}

func BuildTarget(client *ServerFinderConfig) (string, error) {
	for _, r := range rs {
		if client.FinderType == r.Type() {
			return r.BuildTarget(client)
		}
	}
	return "", fmt.Errorf("could not find client finder: %s", client.FinderType)
}
