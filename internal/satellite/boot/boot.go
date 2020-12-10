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
	"errors"
	"os"
	"reflect"
	"sync"
	"syscall"

	"os/signal"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	"github.com/apache/skywalking-satellite/internal/satellite/module"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/plugins"
)

type ModuleContainer map[string][]api.Module

// Start Satellite.
func Start(cfg *config.SatelliteConfig) error {
	log.Init(cfg.Logger)
	plugins.RegisterPlugins()

	// use context to perceive external signal.
	ctx, cancel := context.WithCancel(context.Background())
	addShutdownListener(cancel)

	// boot Satellite
	if modules, err := initModules(cfg); err != nil {
		return err
	} else if err := prepareModules(modules); err != nil {
		return err
	} else {
		bootModules(ctx, modules)
		return nil
	}
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

// initModules init the modules and register the modules to the module container.
func initModules(cfg *config.SatelliteConfig) (ModuleContainer, error) {
	log.Logger.Infof("satellite is initializing...")
	if err := initModuleConfig(cfg); err != nil {
		return nil, err
	}

	// contains the initialized modules.
	container := make(ModuleContainer)
	var sharingClientManager *module.ClientManager
	var sharingNamespace = "sharing"

	// add sharing client manager module
	if cfg.ClientManager != nil {
		sharingClientManager = module.NewClientManager(cfg.ClientManager)
		container[sharingNamespace] = []api.Module{sharingClientManager}
	}

	// add the modules in each namespaces.
	for _, aCfg := range cfg.Agents {
		// the added sequence should follow clientManager, gather, sender and processor to purpose the booting sequence.
		var modules []api.Module
		var usingClientManager = sharingClientManager

		if aCfg.ClientManager != nil {
			usingClientManager = module.NewClientManager(aCfg.ClientManager)
			modules = append(modules, usingClientManager)
		}
		gatherer := module.NewGatherer(aCfg.Gatherer)
		sender := module.NewSender(aCfg.Sender, gatherer, usingClientManager)
		processor := module.NewProcessor(aCfg.Processor, sender, gatherer)

		modules = append(modules, gatherer, sender, processor)
		container[aCfg.ModuleCommonConfig.RunningNamespace] = modules
	}
	return container, nil
}

// initModuleConfig valid the config pattern and inject the common config to the specific module config.
func initModuleConfig(cfg *config.SatelliteConfig) error {
	for _, aCfg := range cfg.Agents {
		if aCfg.Gatherer == nil || aCfg.Sender == nil || aCfg.Processor == nil {
			return errors.New("gatherer, sender, and processor is required in agent config")
		}
		if cfg.ClientManager == nil && aCfg.ClientManager == nil {
			return errors.New("at least one sharing client manager configuration or custom configuration is required")
		}
	}
	// inject module common config to the specific module config
	for _, aCfg := range cfg.Agents {
		aCfg.ClientManager.ModuleCommonConfig = *aCfg.ModuleCommonConfig
		aCfg.Sender.ModuleCommonConfig = *aCfg.ModuleCommonConfig
		aCfg.Gatherer.ModuleCommonConfig = *aCfg.ModuleCommonConfig
		aCfg.Gatherer.ModuleCommonConfig = *aCfg.ModuleCommonConfig
	}
	return nil
}

// prepareModules makes that all modules are in a bootable state.
func prepareModules(container ModuleContainer) error {
	log.Logger.Infof("satellite is prepare to start...")
	var preparedModules []api.Module
	for ns, modules := range container {
		for _, m := range modules {
			preparedModules = append(preparedModules, m)
			if err := m.Prepare(); err != nil {
				for _, preparedModule := range preparedModules {
					preparedModule.Shutdown()
				}
				log.Logger.Errorf("%s module of %s namespace is error in preparing stage, error is %v", reflect.TypeOf(m).String(), ns, err)
				return err
			}
		}
	}
	return nil
}

// bootModules boot all modules.
func bootModules(ctx context.Context, container ModuleContainer) {
	log.Logger.Infof("satellite is starting...")
	var wg sync.WaitGroup
	for _, modules := range container {
		for _, m := range modules {
			m := m
			go func() {
				defer wg.Done()
				wg.Add(1)
				m.Boot(ctx)
			}()
		}
	}
	wg.Wait()
}
