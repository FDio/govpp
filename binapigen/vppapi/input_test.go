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
	"testing"
)

func Test_runCommandInRepo(t *testing.T) {
	type args struct {
		repo    string
		commit  string
		command string
		args    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "branch",
			args: args{
				repo:    "https://github.com/FDio/vpp.git",
				commit:  "master",
				command: "bash",
				args:    []string{"-exc", "bash ./src/scripts/version; make json-api-files && ls ./build-root/install-vpp-native/vpp/share/vpp/api"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//, tt.args.command, tt.args.args...
			_, err := cloneRepoLocally(tt.args.repo, tt.args.commit, "", 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("cloneRepoLocally() error = %v, wantErr %v", err, tt.wantErr)
			}
			//t.Logf("OUTPUT:")
		})
	}
}
