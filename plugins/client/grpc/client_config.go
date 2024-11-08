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

package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/apache/skywalking-satellite/plugins/client/grpc/lb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip" // for install the "gzip" decompressor
	"google.golang.org/grpc/metadata"
)

// loadConfig use the client params to build the grpc client config.
func (c *Client) loadConfig() (*[]grpc.DialOption, error) {
	options := make([]grpc.DialOption, 0)

	if c.EnableTLS {
		configTLS, err := c.configTLS()
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(configTLS)))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	var authHeader metadata.MD
	if c.Authentication != "" {
		authHeader = metadata.New(map[string]string{"Authentication": c.Authentication})
	}

	unaryRequestTimeout, err := time.ParseDuration(c.Timeout.Unary)
	if err != nil {
		return nil, fmt.Errorf("cannot parse the unary request timeout: %v", err)
	}
	streamRequestTimeout, err := time.ParseDuration(c.Timeout.Stream)
	if err != nil {
		return nil, fmt.Errorf("cannot parse the stream request timeout: %v", err)
	}

	// append auth or report error
	options = append(options, grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc,
		cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if authHeader != nil {
			ctx = metadata.NewOutgoingContext(ctx, authHeader)
		}
		timeout, timeoutFunc := context.WithTimeout(ctx, streamRequestTimeout)
		clientStream, err := streamer(timeout, desc, cc, method, opts...)
		if err != nil {
			timeoutFunc()
			c.reportError(err)
			return nil, err
		}
		streamWrapper := &timeoutClientStream{clientStream, timeoutFunc}
		return streamWrapper, err
	}))
	grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if authHeader != nil {
			ctx = metadata.NewOutgoingContext(ctx, authHeader)
		}
		timeout, timeoutFunc := context.WithTimeout(ctx, unaryRequestTimeout)
		defer timeoutFunc()
		err := invoker(timeout, method, req, reply, cc, opts...)
		if err != nil {
			c.reportError(err)
		}
		return err
	})

	// using self build load balancer
	options = append(options, grpc.WithDefaultServiceConfig(fmt.Sprintf("{\"loadBalancingPolicy\":%q}", lb.Name)))

	return &options, nil
}

type timeoutClientStream struct {
	grpc.ClientStream
	timeoutFunc context.CancelFunc
}

func (t *timeoutClientStream) RecvMsg(m interface{}) error {
	//defer t.timeoutFunc()
	return t.ClientStream.RecvMsg(m)
}

// configTLS loads and parse the TLS configs.
func (c *Client) configTLS() (tc *tls.Config, tlsErr error) {
	if err := checkTLSFile(c.CaPemPath); err != nil {
		return nil, err
	}
	tlsConfig := new(tls.Config)
	tlsConfig.Renegotiation = tls.RenegotiateNever
	tlsConfig.InsecureSkipVerify = c.InsecureSkipVerify
	caPem, err := os.ReadFile(c.CaPemPath)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		return nil, fmt.Errorf("failed to append certificates")
	}
	tlsConfig.RootCAs = certPool

	if c.ClientKeyPath != "" && c.ClientPemPath != "" {
		if err := checkTLSFile(c.ClientKeyPath); err != nil {
			return nil, err
		}
		if err := checkTLSFile(c.ClientPemPath); err != nil {
			return nil, err
		}
		clientPem, err := tls.LoadX509KeyPair(c.ClientPemPath, c.ClientKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientPem}
	}
	return tlsConfig, nil
}

// checkTLSFile checks the TLS files.
func checkTLSFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return fmt.Errorf("the TLS file is illegal: %s", path)
	}
	return nil
}
