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
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/apache/skywalking-satellite/plugins/queue/partition"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins"
	fetcher_api "github.com/apache/skywalking-satellite/plugins/fetcher/api"
	forwarder_api "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	receiver_api "github.com/apache/skywalking-satellite/plugins/receiver/api"

	"golang.org/x/mod/modfile"
)

const (
	topLevel       = "# "
	secondLevel    = "## "
	lf             = "\n"
	yamlQuoteStart = "```yaml"
	yamlQuoteEnd   = "```"
	markdownSuffix = ".md"

	commentPrefix = "/ "

	categoryQueue = "Queue"
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

	// remove path
	pluginCatalog.Path = ""

	// rebuild all plugins
	var allPlugins []*Catalog
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
				Name: strings.ReplaceAll(pluginName, "-", " "),
				Path: strings.TrimRight(fmt.Sprintf("%s/%s", pluginFilePath, getPluginDocFileName(category, pluginName)), markdownSuffix),
			})
		}
		curPlugin.Catalog = implements

		if len(implements) > 0 {
			allPlugins = append(allPlugins, curPlugin)
		}
	}
	pluginCatalog.Catalog = allPlugins

	return menu.Save(menuFile)
}

func generatePluginListDoc(docDir string, categories []reflect.Type) error {
	fileName := docDir + "/" + "plugin-list" + markdownSuffix
	docStr := topLevel + "Plugin List" + lf
	for _, category := range categories {
		docStr += "- " + category.Name() + lf
		pluginList := getPluginsByCategory(category)
		for _, pluginName := range pluginList {
			docStr += "	- [" + pluginName + "](./" + getPluginDocFileName(category, pluginName) + ")" + lf
			if err := generatePluginDoc(docDir, category, pluginName); err != nil {
				return err
			}
		}
	}
	return writeDoc([]byte(docStr), fileName)
}

func generatePluginDoc(docDir string, category reflect.Type, pluginName string) error {
	docFileName := docDir + "/" + getPluginDocFileName(category, pluginName)
	p := plugin.Get(category, plugin.Config{plugin.NameField: pluginName})
	docRes := topLevel + category.Name() + "/" + pluginName + lf
	docRes += secondLevel + "Description" + lf
	docRes += p.Description() + lf
	docRes += generateSupportForwarders(category, p)
	docRes += secondLevel + "DefaultConfig" + lf
	docRes += yamlQuoteStart + generateDefaultConfig(category, p) + yamlQuoteEnd + lf
	docRes += secondLevel + "Configuration" + lf
	docRes += generateConfiguration(category, p) + lf
	return writeDoc([]byte(docRes), docFileName)
}

func GetModuleName() string {
	goModBytes, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return ""
	}

	modName := modfile.ModulePath(goModBytes)
	return modName
}

func generateDefaultConfig(category reflect.Type, p plugin.Plugin) string {
	configs := p.DefaultConfig()
	if category.Name() == categoryQueue {
		partitionQueue := &partition.PartitionedQueue{}
		configs = fmt.Sprintf("%s%s", configs, partitionQueue.DefaultConfig())
	}
	return configs
}

func generateConfiguration(category reflect.Type, p plugin.Plugin) string {
	var content = ""

	content += "|Name|Type|Description|" + lf
	content += "|----|----|-----------|" + lf

	configurations := getConfigurations(category, reflect.TypeOf(p).Elem())
	eachConfigurationItem(configurations, "", func(name, dataType, desc string) {
		content += fmt.Sprintf("| %s | %s | %s |%s", name, dataType, desc, lf)
	})
	if category.Name() == categoryQueue {
		configurations := getConfigurations(category, reflect.TypeOf(&partition.PartitionedQueue{}).Elem())
		eachConfigurationItem(configurations, "", func(name, dataType, desc string) {
			content += fmt.Sprintf("| %s | %s | %s |%s", name, dataType, desc, lf)
		})
	}

	return content
}

func eachConfigurationItem(items []*pluginConfigurationItem, parentName string, consumer func(name, dataType, desc string)) {
	for _, conf := range items {
		consumer(parentName+conf.name, conf.dataType, conf.description)
		eachConfigurationItem(conf.children, parentName+conf.name+".", consumer)
	}
}

type pluginConfigurationItem struct {
	name        string
	description string
	dataType    string
	children    []*pluginConfigurationItem
}

type pluginChildrenFinder struct {
	childType reflect.Type
	squash    bool
}

func getConfigurations(category, p reflect.Type) []*pluginConfigurationItem {
	pluginDir := strings.TrimPrefix(p.PkgPath(), GetModuleName())
	fset := token.NewFileSet()

	d, err := parser.ParseDir(fset, "."+pluginDir, nil, parser.ParseComments)
	if err != nil {
		log.Logger.Warnf("failed to generate plugin [%s] configuration, error: %v", category.Name()+"/"+p.Name(), err)
		return make([]*pluginConfigurationItem, 0)
	}

	result := make([]*pluginConfigurationItem, 0)
	for _, f := range d {
		pack := doc.New(f, "./", 0)
		for _, t := range pack.Types {
			if t.Name != p.Name() {
				continue
			}

			for _, spec := range t.Decl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				for _, field := range structType.Fields.List {
					item, childFinder := parsePluginConfigurationItem(field, p)
					if childFinder != nil {
						configurations := getConfigurations(category, childFinder.childType)
						if childFinder.squash {
							result = append(result, configurations...)
						} else if item != nil {
							item.children = configurations
						}
					}
					if item != nil {
						result = append(result, item)
					}
				}
			}
		}
	}
	return result
}

// parse field to configuration item
func parsePluginConfigurationItem(field *ast.Field, pType reflect.Type) (*pluginConfigurationItem, *pluginChildrenFinder) {
	if field.Tag == nil {
		return nil, nil
	}

	var fieldName = ""
	if field.Names != nil {
		for _, n := range field.Names {
			fieldName += n.Name
		}
	} else {
		expr, ok := field.Type.(*ast.SelectorExpr)
		if !ok {
			return nil, nil
		}
		fieldName = expr.Sel.Name
	}

	pluginField, find := pType.FieldByName(fieldName)
	if !find {
		return nil, nil
	}
	mapStructureValue := pluginField.Tag.Get("mapstructure")
	var confName string
	var childrenFinder *pluginChildrenFinder
	if index := strings.Index(mapStructureValue, ","); index != -1 {
		if strings.Contains(mapStructureValue[index+1:], "squash") {
			if pluginField.Type.Kind() == reflect.Struct {
				return nil, &pluginChildrenFinder{childType: pluginField.Type, squash: true}
			}
			log.Logger.Warnf("Could not identity plugin field: %v", pluginField)
			return nil, nil
		}
		confName = mapStructureValue[:index]
	} else if len(mapStructureValue) > 0 {
		confName = mapStructureValue
	} else {
		confName = fieldName
	}

	var dataType = pluginField.Type.String()
	switch pluginField.Type.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		if pluginField.Type.Elem().PkgPath() != "" {
			childrenFinder = &pluginChildrenFinder{childType: pluginField.Type.Elem()}
		}
	}

	return &pluginConfigurationItem{
		name:        confName,
		dataType:    dataType,
		description: buildPluginDescription(field),
	}, childrenFinder
}

func buildPluginDescription(field *ast.Field) string {
	var comments = ""
	for _, group := range []*ast.CommentGroup{field.Doc, field.Comment} {
		if group != nil {
			for _, comment := range group.List {
				comments += strings.TrimLeft(comment.Text, commentPrefix)
			}
		}
	}
	return comments
}

func generateSupportForwarders(category reflect.Type, p plugin.Plugin) string {
	var forwarders []forwarder_api.Forwarder
	if category.Name() == "Receiver" {
		forwarders = p.(receiver_api.Receiver).SupportForwarders()
	} else if category.Name() == "Fetcher" {
		forwarders = p.(fetcher_api.Fetcher).SupportForwarders()
	}
	if len(forwarders) == 0 {
		return ""
	}
	result := secondLevel + "Support Forwarders" + lf
	for _, forwarder := range forwarders {
		result += " - [" + forwarder.Name() + "](" + getPluginDocFileName(reflect.TypeOf(forwarder).Elem(), forwarder.Name()) + ")" + lf
	}
	return result
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

func writeDoc(docBytes []byte, docFileName string) error {
	if err := ioutil.WriteFile(docFileName, docBytes, os.ModePerm); err != nil {
		return fmt.Errorf("cannot init the plugin doc: %v", err)
	}
	return nil
}

func createDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return os.Mkdir(path, os.ModePerm)
	}
	return err
}
