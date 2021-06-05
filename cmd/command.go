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
	"time"

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
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				EnvVars: []string{"SATELLITE_CONFIG"},
				Value:   "configs/satellite_config.yaml",
			},
			&cli.StringFlag{
				Name:    "shutdown_hook_time",
				Aliases: []string{"t"},
				Usage:   "The hook `TIME` for graceful shutdown, and the time unit is seconds.",
				EnvVars: []string{"SATELLITE_SHUTDOWN_HOOK_TIME"},
				Value:   "5",
			},
		},
		Action: func(c *cli.Context) error {
			configPath := c.String("config")
			shutdownHookTime := time.Second * time.Duration(c.Int("shutdown_hook_time"))
			cfg := config.Load(configPath)
			return boot.Start(cfg, shutdownHookTime)
		},
	}

	cmdDocs = cli.Command{
		Name:  "docs",
		Usage: "generate satellite plugin docs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "The document output root path",
				EnvVars: []string{"SATELLITE_DOC_PATH"},
				Value:   "docs",
			},
			&cli.StringFlag{
				Name:    "menu",
				Aliases: []string{"m"},
				Usage:   "The menu file path",
				EnvVars: []string{"SATELLITE_MENU_PATH"},
				Value:   "/menu.yml",
			},
			&cli.StringFlag{
				Name:    "plugins",
				Aliases: []string{"p"},
				Usage:   "The plugin list dir",
				EnvVars: []string{"SATELLITE_PLUGIN_PATH"},
				Value:   "/plugins",
			},
		},
		Action: func(c *cli.Context) error {
			outputRootPath := c.String("output")
			menuFilePath := c.String("menu")
			pluginFilePath := c.String("plugins")
			return tools.GeneratePluginDoc(outputRootPath, menuFilePath, pluginFilePath)
		},
	}
)
