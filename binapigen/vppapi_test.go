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

package binapigen

import (
	"fmt"
	"testing"
)

func TestCleanMessageComment(t *testing.T) {
	tests := []struct {
		comment  string
		expected string
	}{
		{
			comment: `/** \\brief Test comment.
    @param foo - first parameter
    @param bar - second parameter
*/`,
			expected: `Test comment.
    - foo - first parameter
    - bar - second parameter`,
		},
		{
			comment: `/** \\brief Another test.
*/`,
			expected: "Another test.",
		},
		{
			comment: `/** \\brief Empty param.
    @param 
*/`,
			expected: `Empty param.
    -`,
		},
		{
			comment: `/** \\brief Show the current system timestamp.
    @param client_index - opaque cookie to identify the sender
    @param context - sender context, to match reply w/ request
*/`,
			expected: `Show the current system timestamp.
    - client_index - opaque cookie to identify the sender
    - context - sender context, to match reply w/ request`,
		},
		{
			comment:  "/** \\brief Reply to get the plugin version\n    @param context - returned sender context, to match reply w/ request\n    @param major - Incremented every time a known breaking behavior change is introduced\n    @param minor - Incremented with small changes, may be used to avoid buggy versions\n*/",
			expected: "Reply to get the plugin version\n    - context - returned sender context, to match reply w/ request\n    - major - Incremented every time a known breaking behavior change is introduced\n    - minor - Incremented with small changes, may be used to avoid buggy versions",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			result := CleanMessageComment(test.comment)
			if result != test.expected {
				t.Errorf("Expected:\n%s\nbut got:\n%s", test.expected, result)
			}
		})
	}
}

func TestStripMessageCommentFields(t *testing.T) {
	tests := []struct {
		comment  string
		expected string
	}{
		{
			comment: `    - context - sender context, to match reply w/ request
    - retval - return value for request
    - sw_if_index - software index for the new af_xdp interface`,
			expected: `    - retval - return value for request
    - sw_if_index - software index for the new af_xdp interface`,
		},
		{
			comment:  "Reply to get the plugin version\n    - context - returned sender context, to match reply w/ request\n    - major - Incremented every time a known breaking behavior change is introduced\n    - minor - Incremented with small changes, may be used to avoid buggy versions",
			expected: "Reply to get the plugin version\n    - major - Incremented every time a known breaking behavior change is introduced\n    - minor - Incremented with small changes, may be used to avoid buggy versions",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			result := StripMessageCommentFields(test.comment, "context", "client_index")
			if result != test.expected {
				t.Errorf("Expected:\n%q\nbut got:\n%q", test.expected, result)
			}
		})
	}
}
