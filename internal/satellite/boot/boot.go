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

package boot

import (
	"context"
	"os"
	"sync"
	"syscall"

	"os/signal"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	"github.com/apache/skywalking-satellite/internal/satellite/module"
)

// Start Satellite.
func Start(cfg *config.SatelliteConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	addShutdownListener(cancel)
	initLogger(cfg)
	initModules(cfg)
	if err := prepareModules(); err != nil {
		return err
	}
	bootModules(ctx)
	return nil
}

// addShutdownListener add a close signal listener.
func addShutdownListener(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		cancel()
	}()
}

// initLogger init the global logger.
func initLogger(cfg *config.SatelliteConfig) {
	log.Init(cfg.Logger)
}

// initModules init the modules and register the modules to the module container.
func initModules(cfg *config.SatelliteConfig) {
	log.Logger.Infof("satellite is initializing...")
	for _, agentConfig := range cfg.Agents {
		if agentConfig.ClientManager != nil {
			module.NewModuleService(agentConfig.ClientManager)
		}
		if agentConfig.Sender != nil {
			module.NewModuleService(agentConfig.Sender)
		}
		if agentConfig.Processor != nil {
			module.NewModuleService(agentConfig.Processor)
		}
		if agentConfig.Gatherer != nil {
			module.NewModuleService(agentConfig.Gatherer)
		}
	}
}

// prepareModules makes that all modules are in a bootable state.
func prepareModules() error {
	log.Logger.Infof("satellite is prepare to start...")
	for _, service := range module.GetModuleContainer() {
		if err := service.Prepare(); err != nil {
			log.Logger.Errorf("%s module of %s namespace is error in preparing stage, error is %v", service.Name(), service.Config().NameSpace(), err)
			return err
		}
	}
	return nil
}

// prepareModules boot all modules.
func bootModules(ctx context.Context) {
	log.Logger.Infof("satellite is starting...")
	var wg sync.WaitGroup
	for _, service := range module.GetModuleContainer() {
		service := service
		go func() {
			defer wg.Done()
			wg.Add(1)
			service.Boot(ctx)
		}()
	}
	wg.Wait()
}
