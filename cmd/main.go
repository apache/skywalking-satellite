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
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	_ "go.uber.org/automaxprocs"
)

// version will be initialized when building
var version = "latest"

func main() {
	app := cli.NewApp()
	app.Name = "SkyWalking-Satellite"
	app.Version = version
	app.Compiled = time.Now()
	app.Usage = "Satellite is for collecting APM data."
	app.Description = "A lightweight collector/sidecar could be deployed closing to the target monitored system, to collect metrics, traces, and logs."
	app.Commands = []*cli.Command{
		&cmdStart,
		&cmdDocs,
	}
	app.Action = cli.ShowAppHelp
	if err := app.Run(os.Args); err != nil {
		log.Fatalln("start SkyWalking Satellite fail", err)
	}
}
