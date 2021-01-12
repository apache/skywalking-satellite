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

package main

import (
	"github.com/urfave/cli/v2"

	"github.com/apache/skywalking-satellite/internal/satellite/boot"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	"github.com/apache/skywalking-satellite/internal/satellite/tools"
)

var (
	cmdStart = cli.Command{
		Name:  "start",
		Usage: "start satellite",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FILE`",
				EnvVars: []string{"SATELLITE_CONFIG"},
				Value:   "configs/satellite_config.yaml",
			},
		},
		Action: func(c *cli.Context) error {
			configPath := c.String("config")
			cfg := config.Load(configPath)
			return boot.Start(cfg)
		},
	}

	cmdDocs = cli.Command{
		Name:  "docs",
		Usage: "generate satellite plugin docs",
		Action: func(c *cli.Context) error {
			return tools.GeneratePluginDoc()
		},
	}
)
