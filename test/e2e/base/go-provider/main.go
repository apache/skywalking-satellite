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
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/apache/skywalking-go"
)

func main() {
	// Keep the runtime busy so CPU profile sampling captures observable stacks.
	go func() {
		for {
			doWork()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	engine := gin.New()

	engine.Handle("GET", "/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	engine.Handle("GET", "/profile", func(c *gin.Context) {
		log.Printf("=== /profile called ===")
		doWork()
		c.String(200, "Profiling completed")
	})

	_ = engine.Run(":8080")
}

func doWork() {
	start := time.Now()
	for time.Since(start) < 10*time.Second {
		_ = 1
		for i := 0; i < 1e6; i++ {
			_ = i * i
		}
	}
	log.Printf("doWork() completed after %v", time.Since(start))
}
