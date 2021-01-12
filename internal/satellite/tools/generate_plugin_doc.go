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

package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins"
)

const (
	docDir      = "docs"
	docPath     = docDir + "/plugin-description.md"
	topLevel    = "# "
	SecondLevel = "## "
	thirdLevel  = "### "
	LF          = "\n"
	codeQuote   = "```"
	descStr     = "description"
	confStr     = "defaultConfig"
)

func GeneratePluginDoc() error {
	log.Init(&log.LoggerConfig{})
	plugins.RegisterPlugins()
	// the generated doc content
	var doc string
	// sort categories by dictionary sequence
	var categories []reflect.Type
	for c := range plugin.Reg {
		categories = append(categories, c)
	}
	sort.Slice(categories, func(i, j int) bool {
		return strings.Compare(categories[i].String(), categories[j].String()) <= 0
	})
	for _, category := range categories {
		mapping := plugin.Reg[category]
		var keys []string
		for k := range mapping {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			p := plugin.Get(category, plugin.Config{plugin.NameField: key})
			doc += topLevel + category.String() + LF
			doc += SecondLevel + key + LF
			doc += thirdLevel + descStr + LF + codeQuote + p.Description() + codeQuote + LF
			doc += thirdLevel + confStr + LF + codeQuote + p.DefaultConfig() + codeQuote + LF
		}
	}
	if err := createDir(docDir); err != nil {
		return fmt.Errorf("create docs dir error: %v", err)
	}
	if err := ioutil.WriteFile(docPath, []byte(doc), os.ModePerm); err != nil {
		return fmt.Errorf("cannot init the plugin doc: %v", err)
	}
	log.Logger.Info("Successfully generate documentation!")
	return nil
}

func createDir(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return os.Mkdir(docDir, os.ModePerm)
	}
	return err
}
