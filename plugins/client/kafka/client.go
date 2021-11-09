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

package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/plugins/client/api"
)

const (
	Name     = "kafka-client"
	ShowName = "Kafka Client"
)

type Client struct {
	config.CommonFields
	Brokers            string `mapstructure:"brokers"`              // The Kafka broker addresses (default `localhost:9092`).
	Version            string `mapstructure:"version"`              // The version should follow this pattern, which is `major.minor.veryMinor.patch`.
	EnableTLS          bool   `mapstructure:"enable_TLS"`           // The TLS switch (default false).
	ClientPemPath      string `mapstructure:"client_pem_path"`      // The file path of client.pem. The config only works when opening the TLS switch.
	ClientKeyPath      string `mapstructure:"client_key_path"`      // The file path of client.key. The config only works when opening the TLS switch.
	CaPemPath          string `mapstructure:"ca_pem_path"`          // The file path oca.pem. The config only works when opening the TLS switch.
	RequiredAcks       int16  `mapstructure:"required_acks"`        // 0 means NoResponse, 1 means WaitForLocal and -1 means WaitForAll (default 1).
	ProducerMaxRetry   int    `mapstructure:"producer_max_retry"`   // The producer max retry times (default 3).
	MetaMaxRetry       int    `mapstructure:"meta_max_retry"`       // The meta max retry times (default 3).
	RetryBackoff       int    `mapstructure:"retry_backoff"`        // How long to wait for the cluster to settle between retries (default 100ms).
	MaxMessageBytes    int    `mapstructure:"max_message_bytes"`    // The max message bytes.
	IdempotentWrites   bool   `mapstructure:"idempotent_writes"`    // Ensure that exactly one copy of each message is written when is true.
	ClientID           string `mapstructure:"client_id"`            // A user-provided string sent with every request to the brokers.
	CompressionCodec   int    `mapstructure:"compression_codec"`    // Represents the various compression codecs recognized by Kafka in messages.
	RefreshPeriod      int    `mapstructure:"refresh_period"`       // How frequently to refresh the cluster metadata.
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"` // Controls whether a client verifies the server's certificate chain and host name.

	// components
	client    sarama.Client // The kafka producer.
	listeners []chan<- api.ClientStatus
	status    api.ClientStatus
	ctx       context.Context    // Parent ctx
	cancel    context.CancelFunc // Parent ctx cancel function
}

func (c *Client) Name() string {
	return Name
}

func (c *Client) ShowName() string {
	return ShowName
}

func (c *Client) Description() string {
	return "The Kafka client is a sharing plugin to keep connection with the Kafka brokers and delivery the data to it."
}

func (c *Client) DefaultConfig() string {
	return `
# The Kafka broker addresses (default localhost:9092). Multiple values are separated by commas.
brokers: localhost:9092

# The Kafka version should follow this pattern, which is major_minor_veryMinor_patch (default 1.0.0.0).
version: 1.0.0.0

# The TLS switch (default false).
enable_TLS: false

# The file path of client.pem. The config only works when opening the TLS switch.
client_pem_path: ""

# The file path of client.key. The config only works when opening the TLS switch.
client_key_path: ""

# The file path oca.pem. The config only works when opening the TLS switch.
ca_pem_path: ""

# 0 means NoResponse, 1 means WaitForLocal and -1 means WaitForAll (default 1).
required_acks: 1

# The producer max retry times (default 3).
producer_max_retry: 3

# The meta max retry times (default 3).
meta_max_retry: 3

# How long to wait for the cluster to settle between retries (default 100ms). Time unit is ms.
retry_backoff: 100

# The max message bytes.
max_message_bytes: 1000000

# If enabled, the producer will ensure that exactly one copy of each message is written (default false).
idempotent_writes: false

# A user-provided string sent with every request to the brokers for logging, debugging, and auditing purposes (default Satellite).
client_id: Satellite

# Compression codec represents the various compression codecs recognized by Kafka in messages. 0 : None, 1 : Gzip, 2 : Snappy, 3 : LZ4, 4 : ZSTD
compression_codec: 0

# How frequently to refresh the cluster metadata in the background. Defaults to 10 minutes. The unit is minute.
refresh_period: 10

# InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
insecure_skip_verify: true
`
}

func (c *Client) Prepare() error {
	cfg, err := c.loadConfig()
	if err != nil {
		return fmt.Errorf("cannot init the kafka producer: %v", err)
	}
	sarama.Logger = log.Logger
	client, err := sarama.NewClient(strings.Split(c.Brokers, ","), cfg)
	if err != nil {
		return fmt.Errorf("cannot init the kafka client: %v", err)
	}
	c.client = client
	c.status = api.Connected
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.listeners = make([]chan<- api.ClientStatus, 0)
	return nil
}

func (c *Client) Close() error {
	c.cancel()
	defer log.Logger.Info("kafka client is closed")
	return c.client.Close()
}

func (c *Client) GetConnectedClient() interface{} {
	return c.client
}

func (c *Client) RegisterListener(listener chan<- api.ClientStatus) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) Start() error {
	// start supported processes.
	go c.snifferBrokerStatus()
	return nil
}
