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

	"gopkg.in/yaml.v3"
)

type Catalog struct {
	Name    string     `yaml:"name,omitempty"`
	Path    string     `yaml:"path,omitempty"`
	Catalog []*Catalog `yaml:"catalog,omitempty"`
}

// LoadCatalog data from file
func LoadCatalog(filename string) (*Catalog, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read the menu file: %v", err)
	}

	catalog := Catalog{}
	err = yaml.Unmarshal(bytes, &catalog)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal menu file: %v", err)
	}
	return &catalog, nil
}

// Find Catalog by paths
func (c *Catalog) Find(namePaths ...string) *Catalog {
	if c.Catalog == nil {
		return nil
	}

	children := c.Catalog
	finded := c
	for _, name := range namePaths {
		finded = nil
		for _, cc := range children {
			if cc.Name == name {
				finded = cc
				break
			}
		}
		if finded == nil {
			return nil
		}
		children = finded.Catalog
	}
	return finded
}

func (c *Catalog) Save(filename string) error {
	content := []byte(`# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

`)

	marshal, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, append(content, marshal...), os.ModePerm); err != nil {
		return fmt.Errorf("cannot write catalog: %v", err)
	}
	return nil
}
