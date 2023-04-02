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

package pprof

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry/none"
)

func init() {
	telemetry.Register("pprof", &Server{}, false)
}

type Server struct {
	svr *http.Server
}

func (s *Server) Start(config *telemetry.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.svr = &http.Server{
		Addr:              config.PProfService.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}
	go func() {
		log.Logger.WithField("addr", config.PProfService.Address).Debugf("start pprof server")
		if err := s.svr.ListenAndServe(); err != nil {
			log.Logger.WithField("addr", config.PProfService.Address).Warnf("starting pprof server failure: %v", err)
		}
	}()
	return nil
}

func (s *Server) AfterSharingStart() error {
	return nil
}

func (s *Server) Close() error {
	return s.svr.Close()
}

func (s *Server) NewCounter(name, help string, labels ...string) telemetry.Counter {
	return &none.Counter{}
}

func (s *Server) NewGauge(name, help string, getter func() float64, labels ...string) telemetry.Gauge {
	return &none.Gauge{}
}

func (s *Server) NewDynamicGauge(name, help string, labels ...string) telemetry.DynamicGauge {
	return &none.DynamicGauge{}
}

func (s *Server) NewTimer(name, help string, labels ...string) telemetry.Timer {
	return &none.Timer{}
}
