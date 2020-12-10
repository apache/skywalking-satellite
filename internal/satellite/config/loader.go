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

package config

import (
	"bytes"
	"fmt"
	"sync"

	"io/ioutil"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	cfgLock sync.Mutex
)

// Load SatelliteConfig. The func could not use global logger of Satellite, because it is executed before logger initialization.
func Load(configPath string) *SatelliteConfig {
	cfgLock.Lock()
	defer cfgLock.Unlock()
	fmt.Printf("load config from : %s\n", configPath)
	cfg, err := load(configPath)
	if err != nil {
		panic(fmt.Errorf("could not load config form the path: %s, the error is :%v", configPath, err))
	} else {
		return cfg
	}
}

// load SatelliteConfig from the yaml config.
func load(configPath string) (*SatelliteConfig, error) {
	absolutePath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigType("yaml")
	cfg := SatelliteConfig{}
	if err := v.ReadConfig(bytes.NewReader(content)); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
