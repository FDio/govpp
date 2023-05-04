//  Copyright (c) 2023 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

const modulePath = "go.fd.io/govpp"

func init() {
	formatter := &logrus.TextFormatter{
		EnvironmentOverrideColors: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			call := strings.TrimPrefix(frame.Function, modulePath)
			function = fmt.Sprintf("%s()", strings.TrimPrefix(call, "/"))
			_, file = filepath.Split(frame.File)
			file = fmt.Sprintf("%s:%d", file, frame.Line)
			return color.Debug.Sprint(function), color.Secondary.Sprint(file)
		},
	}
	logrus.SetFormatter(formatter)
	logrus.AddHook(&callerHook{})
}

type callerHook struct {
}

func (c *callerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (c *callerHook) Fire(entry *logrus.Entry) error {
	if fn, ok := entry.Data[logrus.FieldKeyFunc]; ok && entry.Caller != nil {
		fmt.Printf("LOG ENTRY (fn: %v): %+v\n", fn, entry)
	}
	return nil
}
