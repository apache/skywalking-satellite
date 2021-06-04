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
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	"github.com/apache/skywalking-satellite/plugins/client/api"
)

// sniffer
func (c *Client) snifferChannelStatus() {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	timeTicker := time.NewTicker(time.Duration(c.CheckPeriod) * time.Second)
	for {
		select {
		case <-timeTicker.C:
			state := c.client.GetState()
			if state == connectivity.Shutdown || state == connectivity.TransientFailure {
				c.updateStatus(api.Disconnect)
			} else if state == connectivity.Ready || state == connectivity.Idle {
				c.updateStatus(api.Connected)
			}
		case <-ctx.Done():
			timeTicker.Stop()
			return
		}
	}
}

func (c *Client) reportError(err error) {
	if err == nil {
		return
	}
	fromError, ok := status.FromError(err)
	if ok {
		errCode := fromError.Code()
		if errCode == codes.Unavailable || errCode == codes.PermissionDenied ||
			errCode == codes.Unauthenticated || errCode == codes.ResourceExhausted || errCode == codes.Unknown {
			c.updateStatus(api.Disconnect)
		}
	}
}

func (c *Client) updateStatus(clientStatus api.ClientStatus) {
	if c.status != clientStatus {
		c.status = clientStatus
		for _, listener := range c.listeners {
			listener <- c.status
		}
	}
}
