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
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_overrideString(t *testing.T) {
	type args struct {
		expression string
		env        string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "override_string",
			args: args{
				expression: "${TEST_OVERRIDE_STRING:test_str}",
				env:        "TEST_OVERRIDE_STRING=test_override_string",
			},
			want: "test_override_string",
		},
		{
			name: "no_override_string",
			args: args{
				expression: "${TEST_NO_OVERRIDE_STRING:test_str}",
			},
			want: "test_str",
		},
		{
			name: "no_override_false",
			args: args{
				expression: "${TEST_NO_OVERRIDE_FALSE:false}",
			},
			want: false,
		},
		{
			name: "no_override_true",
			args: args{
				expression: "${TEST_NO_OVERRIDE_TRUE:true}",
			},
			want: true,
		},
		{
			name: "override_boolean",
			args: args{
				expression: "${TEST_OVERRIDE_BOOLEAN:true}",
				env:        "TEST_OVERRIDE_BOOLEAN=false",
			},
			want: false,
		},
		{
			name: "no_override_int",
			args: args{
				expression: "${TEST_OVERRIDE_INT:10}",
			},
			want: 10,
		},
		{
			name: "override_int",
			args: args{
				expression: "${TEST_OVERRIDE_INT:10}",
				env:        "TEST_OVERRIDE_INT=15",
			},
			want: 15,
		},
		{
			name: "override_float",
			args: args{
				expression: "${TEST_OVERRIDE_FLOAT:10.5}",
				env:        "TEST_OVERRIDE_FLOAT=15.7",
			},
			want: 15.7,
		},
		{
			name: "no_override_force_string",
			args: args{
				expression: "${TEST_NO_OVERRIDE_FORCE_STRING:\"10.5\"}",
			},
			want: "10.5",
		},
	}
	for _, tt := range tests {
		if tt.args.env != "" {
			envArr := strings.Split(tt.args.env, "=")
			if err := os.Setenv(envArr[0], envArr[1]); err != nil {
				t.Fatalf("cannot set the env %s  config: %v", tt.args.env, err)
			}
		}
		regex := regexp.MustCompile(RegularExpression)
		t.Run(tt.name, func(t *testing.T) {
			if got := overrideString(tt.args.expression, regex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("overrideString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_overrideMapStringInterface(t *testing.T) {
	type args struct {
		cfg map[string]interface{}
		env map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "test-overrideByEnv",
			args: args{
				cfg: map[string]interface{}{
					"stringKey": "${OVERRIDE_STRING_KEY:stringKey}",
					"intKey":    "${OVERRIDE_INT_KEY:10}",
					"boolKey":   "${OVERRIDE_BOOL_KEY:false}",
					"mapKey": map[string]interface{}{
						"mapStringKey": "${OVERRIDE_STRING_KEY:stringKey}",
						"mapIntKey":    "${OVERRIDE_INT_KEY:10}",
						"mapBoolKey":   "${OVERRIDE_BOOL_KEY:false}",
						"mapInterfaceKey": map[interface{}]interface{}{
							"mapinterfaceStringKey": "${OVERRIDE_STRING_KEY:stringKey}",
							"mapinterfaceIntKey":    "${OVERRIDE_INT_KEY:10}",
							"mapinterfaceBoolKey":   "${OVERRIDE_BOOL_KEY:false}",
						},
					},
				},
				env: map[string]string{
					"OVERRIDE_STRING_KEY": "env-string",
					"OVERRIDE_INT_KEY":    "100",
					"OVERRIDE_BOOL_KEY":   "true",
				},
			},
			want: map[string]interface{}{
				"stringKey": "env-string",
				"intKey":    100,
				"boolKey":   true,
				"mapKey": map[string]interface{}{
					"mapStringKey": "env-string",
					"mapIntKey":    100,
					"mapBoolKey":   true,
					"mapInterfaceKey": map[string]interface{}{
						"mapinterfaceStringKey": "env-string",
						"mapinterfaceIntKey":    100,
						"mapinterfaceBoolKey":   true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		for k, v := range tt.args.env {
			err := os.Setenv(k, v)
			if err != nil {
				t.Fatalf("cannot set the env config{%s=%s}: %v", k, v, err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(RegularExpression)
			got := overrideMapStringInterface(tt.args.cfg, regex)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("overrideConfigByEnv()  got = %v, want = %v", got, tt.want)
			}
		})
	}
}
