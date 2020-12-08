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

package test

import (
	"fmt"
	"os"
	"strings"

	"github.com/apache/skywalking-satellite/internal/pkg/constant"
)

func FindRootPath() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not find the project root path: %v", err)
	}
	return pwd[0 : strings.Index(pwd, constant.ProjectName)+len(constant.ProjectName)], nil
}
