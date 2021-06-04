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
	topLevel       = "# "
	secondLevel    = "## "
	lf             = "\n"
	yamlQuoteStart = "```yaml"
	yamlQuoteEnd   = "```"
	markdownSuffix = ".md"
)

func GeneratePluginDoc(outputRootPath, menuFilePath, pluginFilePath string) error {
	log.Init(&log.LoggerConfig{})
	plugins.RegisterPlugins()

	pluginPath := fmt.Sprintf("%s%s", outputRootPath, pluginFilePath)
	if err := createDir(pluginPath); err != nil {
		return fmt.Errorf("create docs dir error: %v", err)
	}
	if err := generatePluginListDoc(pluginPath, getSortedCategories()); err != nil {
		return err
	}
	if err := updateMenuPluginListDoc(outputRootPath, menuFilePath, pluginFilePath, getSortedCategories()); err != nil {
		return err
	}
	log.Logger.Info("Successfully generate documentation!")
	return nil
}

// sort categories by dictionary sequence
func getSortedCategories() []reflect.Type {
	var categories []reflect.Type
	for c := range plugin.Reg {
		categories = append(categories, c)
	}
	sort.Slice(categories, func(i, j int) bool {
		return strings.Compare(categories[i].String(), categories[j].String()) <= 0
	})
	return categories
}

func updateMenuPluginListDoc(outputRootPath, menuFilePath, pluginFilePath string, categories []reflect.Type) error {
	menuFile := fmt.Sprintf("%s%s", outputRootPath, menuFilePath)
	menu, err := LoadCatalog(menuFile)
	if err != nil {
		return err
	}

	// find plugin Catalog
	pluginCatalog := menu.Find("Setup", "Plugins")
	if pluginCatalog == nil {
		return fmt.Errorf("cannot find plugins Catalog")
	}

	// rebuild all plugins
	var plugins []*Catalog
	for _, category := range categories {
		// plugin
		implements := []*Catalog{}
		curPlugin := &Catalog{
			Name: strings.ToLower(category.Name()),
		}

		// all implements
		pluginList := getPluginsByCategory(category)
		for _, pluginName := range pluginList {
			implements = append(implements, &Catalog{
				Name: pluginName,
				Path: strings.TrimRight(fmt.Sprintf("%s/%s", pluginFilePath, getPluginDocFileName(category, pluginName)), markdownSuffix),
			})
		}
		curPlugin.Catalog = implements

		if len(implements) > 0 {
			plugins = append(plugins, curPlugin)
		}
	}
	pluginCatalog.Catalog = plugins

	return menu.Save(menuFile)
}

func generatePluginListDoc(docDir string, categories []reflect.Type) error {
	fileName := docDir + "/" + "plugin-list" + markdownSuffix
	doc := topLevel + "Plugin List" + lf
	for _, category := range categories {
		doc += "- " + category.Name() + lf
		pluginList := getPluginsByCategory(category)
		for _, pluginName := range pluginList {
			doc += "	- [" + pluginName + "](./" + getPluginDocFileName(category, pluginName) + ")" + lf
			if err := generatePluginDoc(docDir, category, pluginName); err != nil {
				return err
			}
		}
	}
	return writeDoc([]byte(doc), fileName)
}

func generatePluginDoc(docDir string, category reflect.Type, pluginName string) error {
	docFileName := docDir + "/" + getPluginDocFileName(category, pluginName)
	p := plugin.Get(category, plugin.Config{plugin.NameField: pluginName})
	doc := topLevel + category.Name() + "/" + pluginName + lf
	doc += secondLevel + "Description" + lf
	doc += p.Description() + lf
	doc += secondLevel + "DefaultConfig" + lf
	doc += yamlQuoteStart + p.DefaultConfig() + yamlQuoteEnd + lf
	return writeDoc([]byte(doc), docFileName)
}

func getPluginsByCategory(category reflect.Type) []string {
	mapping := plugin.Reg[category]
	var keys []string
	for k := range mapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getPluginDocFileName(category reflect.Type, pluginName string) string {
	return strings.ToLower(category.Name() + "_" + pluginName + markdownSuffix)
}

func writeDoc(doc []byte, docFileName string) error {
	if err := ioutil.WriteFile(docFileName, doc, os.ModePerm); err != nil {
		return fmt.Errorf("cannot init the plugin doc: %v", err)
	}
	return nil
}

func createDir(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return os.Mkdir(path, os.ModePerm)
	}
	return err
}
