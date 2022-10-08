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
	"os"
	"reflect"
	"strings"
	"sync"

	"path/filepath"

	"github.com/spf13/viper"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
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

// load SatelliteConfig from the yaml config and override the value with env config.
func load(configPath string) (*SatelliteConfig, error) {
	absolutePath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigType("yaml")
	cfg := NewDefaultSatelliteConfig()
	if err := v.ReadConfig(bytes.NewReader(content)); err != nil {
		return nil, err
	}
	overrideConfigByEnv(v)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	propagateCommonFieldsInSharing(cfg.Sharing)
	propagateCommonFieldsInPipes(cfg.Pipes)
	return cfg, nil
}

// propagateCommonFieldsInPipes propagates the common fields to every module and the dependency plugin.
func propagateCommonFieldsInPipes(pipes []*PipeConfig) {
	for _, pipe := range pipes {
		pipe.Gatherer.CommonFields = pipe.PipeCommonConfig
		pipe.Sender.CommonFields = pipe.PipeCommonConfig
		pipe.Processor.CommonFields = pipe.PipeCommonConfig
		propagateCommonFieldsInStruct(pipe.Gatherer, pipe.PipeCommonConfig)
		propagateCommonFieldsInStruct(pipe.Sender, pipe.PipeCommonConfig)
		propagateCommonFieldsInStruct(pipe.Processor, pipe.PipeCommonConfig)
	}
}

// propagate the common fields to the sharing plugins.
func propagateCommonFieldsInSharing(sharing *SharingConfig) {
	propagateCommonFieldsInStruct(sharing, sharing.SharingCommonConfig)
}

// propagate the common fields to the fields that is one of `plugin.config` or `[]plugin.config` types.
func propagateCommonFieldsInStruct(cfg interface{}, commonFields *config.CommonFields) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.Type().NumField(); i++ {
		fieldVal := v.Field(i).Interface()
		if arr, ok := fieldVal.([]plugin.Config); arr != nil && ok {
			for _, pc := range arr {
				propagateCommonFields(pc, commonFields)
			}
		} else if pc, ok := fieldVal.(plugin.Config); pc != nil && ok {
			propagateCommonFields(pc, commonFields)
		}
	}
}

// propagate the common fields to the `plugin.config`.
func propagateCommonFields(pc plugin.Config, cf *config.CommonFields) {
	v := reflect.ValueOf(cf)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if tagVal := t.Field(i).Tag.Get(config.TagName); tagVal != "" {
			pc[strings.ToLower(config.CommonFieldsName)+"_"+tagVal] = v.Field(i).Interface()
		}
	}
}
