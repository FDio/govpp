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

package vppapi

import (
	"reflect"
	"testing"
)

func TestParseInput(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name      string
		args      args
		want      *Input
		wantErr   bool
		pre, post func()
	}{
		{
			name: "local dir",
			args: args{"/usr/share/vpp/api"},
			want: &Input{
				Format: "dir",
				Path:   "/usr/share/vpp/api",
			},
		},
		{
			name: "git repo",
			args: args{".git#branch=master"},
			want: &Input{
				Format: "git",
				Path:   ".git",
				Options: map[string]string{
					"branch": "master",
				},
			},
		},
		{
			name: "archive",
			args: args{"input.tar.gz"},
			want: &Input{
				Format: "tar",
				Path:   "input.tar.gz",
			},
		},
		{
			name: "remote archive",
			args: args{"https://example.com/input.tar.gz"},
			want: &Input{
				Format: "tar",
				Path:   "https://example.com/input.tar.gz",
			},
		},
		{
			name: "remote repo",
			args: args{"https://github.com/FDio/vpp.git"},
			want: &Input{
				Format: "git",
				Path:   "https://github.com/FDio/vpp.git",
			},
		},
		{
			name: "no .git",
			args: args{"https://github.com/FDio/vpp#format=git"},
			want: &Input{
				Format: "git",
				Path:   "https://github.com/FDio/vpp",
			},
		},
		{
			name: "git tag",
			args: args{"https://github.com/FDio/vpp.git#tag=v23.02"},
			want: &Input{
				Format: "git",
				Path:   "https://github.com/FDio/vpp.git",
				Options: map[string]string{
					"tag": "v23.02",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre()
			}
			if tt.post != nil {
				t.Cleanup(tt.post)
			}
			if tt.want.Options == nil {
				tt.want.Options = map[string]string{}
			}

			got, err := ParseInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveVppInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveVppInput() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
