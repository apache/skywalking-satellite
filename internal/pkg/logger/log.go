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

package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Default logger config.
const (
	defaultLogPattern  = "%time [%level][%field] - %msg"
	defaultTimePattern = "2006-01-02 15:04:05.001"
)

type formatter struct {
	logPattern  string
	timePattern string
}

// Option is a function to set formatter config.
type Option func(f *formatter)

// Log is the global logger.
var Log *logrus.Logger
var once sync.Once

// The Log init method, keep Log as a singleton.
func Init(opts ...Option) {
	once.Do(func() {
		if Log == nil {
			Log = logrus.New()
		}
		Log.SetOutput(os.Stdout)
		Log.SetLevel(logrus.InfoLevel)
		f := &formatter{}
		for _, opt := range opts {
			opt(f)
		}
		if f.logPattern == "" {
			f.logPattern = defaultLogPattern
		}
		if f.timePattern == "" {
			f.timePattern = defaultTimePattern
		}
		if !strings.Contains(f.logPattern, "\n") {
			f.logPattern += "\n"
		}
		Log.SetFormatter(f)
	})
}

// Set the log pattern in formatter.
func SetLogPattern(logPattern string) Option {
	return func(f *formatter) {
		f.logPattern = logPattern
	}
}

// Set the time pattern in formatter.
func SetTimePattern(timePattern string) Option {
	return func(f *formatter) {
		f.timePattern = timePattern
	}
}

// Format supports unified log output format that has %time, %level, %field and %msg.
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.logPattern
	output = strings.Replace(output, "%time", entry.Time.Format(f.timePattern), 1)
	output = strings.Replace(output, "%level", entry.Level.String(), 1)
	output = strings.Replace(output, "%field", buildFields(entry), 1)
	output = strings.Replace(output, "%msg", entry.Message, 1)
	return []byte(output), nil
}

func buildFields(entry *logrus.Entry) string {
	var fields []string
	for key, val := range entry.Data {
		stringVal, ok := val.(string)
		if !ok {
			stringVal = fmt.Sprint(val)
		}
		fields = append(fields, key+"="+stringVal)
	}
	return strings.Join(fields, ",")
}
