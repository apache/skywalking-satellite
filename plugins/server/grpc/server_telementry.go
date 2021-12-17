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
	"strings"

	"google.golang.org/grpc/peer"
)

func GetPeerHostFromStreamContext(ctx context.Context) string {
	peerAddr := GetPeerAddressFromStreamContext(ctx)
	if inx := strings.IndexByte(peerAddr, ':'); inx > 0 {
		peerAddr = peerAddr[:strings.IndexByte(peerAddr, ':')]
	}
	return peerAddr
}

func GetPeerAddressFromStreamContext(ctx context.Context) string {
	if peerAddr, ok := peer.FromContext(ctx); ok {
		return peerAddr.Addr.String()
	}
	return ""
}
