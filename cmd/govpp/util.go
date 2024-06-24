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
	"regexp"
	"sort"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

var (
	clrWhite       = color.Style{color.White}
	clrDiffMessage = color.Style{color.Cyan}
	clrDiffOption  = color.Style{color.Blue}
	clrDiffFile    = color.Style{color.Yellow}
	clrDiffVersion = color.Style{color.LightMagenta}
	clrDiffNumber  = color.Style{color.LightBlue}
)

const (
	codeSuffix = "[0m"
	codeExpr   = `(\\u001b|\033)\[[\d;?]+m`
)

var codeRegex = regexp.MustCompile(codeExpr)

func clearColorCode(str string) string {
	if !strings.Contains(str, codeSuffix) {
		return str
	}
	return codeRegex.ReplaceAllString(str, "")
}

func mapStr(data map[string]string) string {
	var str string
	for k, v := range data {
		if len(str) > 0 {
			str += ", "
		}
		if v == "" {
			str += k
		} else {
			str += fmt.Sprintf("%s: %q", k, v)
		}
	}
	return str
}

func mapStrOrdered(data map[string]string) string {
	var strs []string
	for k, v := range data {
		var str string
		if v == "" {
			str = k
		} else {
			str = fmt.Sprintf("%s: %q", k, v)
		}
		strs = append(strs, str)
	}
	sort.Strings(strs)
	return strings.Join(strs, ", ")
}

func must(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}
