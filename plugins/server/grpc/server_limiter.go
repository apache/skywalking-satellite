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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"

	"github.com/sirupsen/logrus"

	"github.com/shirou/gopsutil/cpu"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

const CPUUpdateInterval = time.Second * 5

type AcceptConnectionConfig struct {
	CPUUtilization  float64 `mapstructure:"cpu_utilization"`  // The max CPU utilization limit
	ConnectionCount int32   `mapstructure:"connection_count"` // The max connection count
}

type AcceptLimiter struct {
	Config           AcceptConnectionConfig
	ActiveConnection int32
	CurrentCPU       float64
	logger           *logrus.Entry

	telemetry.Counter
}

func NewAcceptLimiter(config AcceptConnectionConfig) (*AcceptLimiter, error) {
	limiter := &AcceptLimiter{Config: config}
	if err := limiter.init(); err != nil {
		return nil, err
	}
	return limiter, nil
}

func (a *AcceptLimiter) init() error {
	ctx := context.Background()

	// init logger
	a.logger = log.Logger.
		WithField("client-name", Name).
		WithField("component", "accept-limiter")

	// cpu analyzer adaptor
	if err := a.cpuUsage(ctx); err != nil {
		return fmt.Errorf("could not find cpu usage analyzer: %v", err)
	}

	// init telemetry
	telemetry.NewGauge("grpc_server_cpu_gauge", "The cpu usage of satellite process", func() float64 {
		return a.CurrentCPU
	})
	telemetry.NewGauge("grpc_server_connection_count", "The active connection count of gRPC server", func() float64 {
		return float64(a.ActiveConnection)
	})
	return nil
}

func (a *AcceptLimiter) cpuUsage(ctx context.Context) error {
	if _, err := cpu.Times(false); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if percent, err := cpu.Percent(CPUUpdateInterval, false); err != nil {
					a.logger.Warnf("error query the cpu usage: %v", err)
				} else {
					a.CurrentCPU = percent[0]
				}
			}
		}
	}()
	return nil
}

func (a *AcceptLimiter) CouldHandleConnection() bool {
	// cpu check
	if a.CurrentCPU >= a.Config.CPUUtilization {
		return false
	}

	// active connection check
	if a.ActiveConnection >= a.Config.ConnectionCount {
		return false
	}
	// try to add the active count
	if atomic.AddInt32(&a.ActiveConnection, 1) > a.Config.ConnectionCount {
		atomic.AddInt32(&a.ActiveConnection, -1)
		return false
	}

	return true
}

func (a *AcceptLimiter) CloseConnection() {
	atomic.AddInt32(&a.ActiveConnection, -1)
}
