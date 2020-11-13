// Copyright 2018-2025 JDCLOUD.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import "fmt"

const (
	LogFatal = iota
	LogError
	LogWarn
	LogInfo  
)

type Logger interface {
	Log(level int, message... interface{})
}

type DefaultLogger struct {
	Level int
}

func NewDefaultLogger(level int) *DefaultLogger {
	return &DefaultLogger{level}
}

func (logger DefaultLogger) Log (level int, message... interface{}) {
	if level <= logger.Level {
		fmt.Println(message...)
	}
}

