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

	"github.com/prometheus/common/config"
	"gopkg.in/yaml.v3"
)

type KubernetesConfig struct {
	// The kubernetes API server address, If not define means using in kubernetes mode to connect
	APIServer string `mapstructure:"api_server"`
	// Connect to API Server Config
	HTTPClientConfig HTTPClientConfig `mapstructure:",squash" yaml:",inline"`
	// Support to lookup namespaces
	Namespaces []string `mapstructure:"namespaces"`
	// The kind of api
	Kind string `mapstructure:"kind"`
	// The kind selector
	Selector Selector `mapstructure:"selector"`
	// How to get the address exported port
	ExtraPort ExtraPort `mapstructure:"extra_port"`
}

// HTTPClientConfig configures an HTTP client.
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `mapstructure:"basic_auth" yaml:"basic_auth,omitempty"`
	// The bearer token for the targets.
	BearerToken Secret `mapstructure:"bearer_token" yaml:"bearer_token,omitempty"`
	// The bearer token file for the targets.
	BearerTokenFile string `mapstructure:"bearer_token_file" yaml:"bearer_token_file,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL string `mapstructure:"proxy_url" yaml:"proxy_url,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig TLSConfig `mapstructure:"tls_config" yaml:"tls_config,omitempty"`
}

// BasicAuth contains basic HTTP authentication credentials.
type BasicAuth struct {
	Username     string `mapstructure:"username" yaml:"username"`
	Password     Secret `mapstructure:"password" yaml:"password,omitempty"`
	PasswordFile string `mapstructure:"password_file" yaml:"password_file,omitempty"`
}

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `mapstructure:"ca_file" yaml:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `mapstructure:"cert_file" yaml:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `mapstructure:"key_file" yaml:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `mapstructure:"server_name" yaml:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// Secret special type for storing secrets.
type Secret string

type Selector struct {
	Label string `mapstructure:"label" yaml:"label,omitempty"`
	Field string `mapstructure:"field" yaml:"field,omitempty"`
}

type ExtraPort struct {
	Port int `mapstructure:"port"`
}

// convert config data
func (c *HTTPClientConfig) convertHTTPConfig() (*config.HTTPClientConfig, error) {
	marshal, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not identity the http client config: %v", err)
	}

	out := &config.HTTPClientConfig{}
	if err = yaml.Unmarshal(marshal, out); err != nil {
		return nil, fmt.Errorf("could not convert http client: %v", err)
	}
	return out, nil
}
