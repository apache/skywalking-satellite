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

package log

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
	defaultTimePattern = "2006-01-02 15:04:05.000"
)

// LoggerConfig initializes the global logger config.
type LoggerConfig struct {
	LogPattern  string `mapstructure:"log_pattern"`
	TimePattern string `mapstructure:"time_pattern"`
	Level       string `mapstructure:"level"`
}

// FormatOption is a function to set formatter config.
type FormatOption func(f *formatter)

// ConfigOption is a function to set logger config.
type ConfigOption func(l *logrus.Logger)

type formatter struct {
	logPattern  string
	timePattern string
}

// Logger is the global logger.
var Logger *logrus.Logger
var once sync.Once

func Init(cfg *LoggerConfig) {
	once.Do(func() {
		var configOpts []ConfigOption
		var formatOpts []FormatOption

		if cfg.Level != "" {
			configOpts = append(configOpts, SetLevel(cfg.Level))
		} else {
			configOpts = append(configOpts, SetLevel(logrus.InfoLevel.String()))
		}
		if cfg.TimePattern != "" {
			formatOpts = append(formatOpts, SetTimePattern(cfg.TimePattern))
		} else {
			formatOpts = append(formatOpts, SetTimePattern(defaultTimePattern))
		}
		if cfg.LogPattern != "" {
			formatOpts = append(formatOpts, SetLogPattern(cfg.LogPattern))
		} else {
			formatOpts = append(formatOpts, SetLogPattern(defaultLogPattern))
		}
		initBySettings(configOpts, formatOpts)
	})
}

// The Logger init method, keep Logger as a singleton.
func initBySettings(configOpts []ConfigOption, formatOpts []FormatOption) {
	// Default Logger.
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	for _, opt := range configOpts {
		opt(Logger)
	}
	// Default formatter.
	f := &formatter{}
	for _, opt := range formatOpts {
		opt(f)
	}
	if !strings.Contains(f.logPattern, "\n") {
		f.logPattern += "\n"
	}
	Logger.SetFormatter(f)
}

// Put the log pattern in formatter.
func SetLogPattern(logPattern string) FormatOption {
	return func(f *formatter) {
		f.logPattern = logPattern
	}
}

// Put the time pattern in formatter.
func SetTimePattern(timePattern string) FormatOption {
	return func(f *formatter) {
		f.timePattern = timePattern
	}
}

// Put the time pattern in formatter.
func SetLevel(levelStr string) ConfigOption {
	return func(logger *logrus.Logger) {
		level, err := logrus.ParseLevel(levelStr)
		if err != nil {
			fmt.Printf("logger level does not exist: %s, level would be set info", levelStr)
			level = logrus.InfoLevel
		}
		logger.SetLevel(level)
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
