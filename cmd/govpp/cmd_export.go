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
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

type ExportCmdOptions struct {
	Input  string
	Output string
}

func newExportCmd() *cobra.Command {
	var (
		opts = ExportCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:   "export INPUT",
		Short: "Export input files to output location",
		Long:  "Export the files from the input location to an output location.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			if opts.Input == "" {
				opts.Input = detectVppApiInput()
			}
			return runExportCmd(opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output directory for the exported files")
	must(cobra.MarkFlagRequired(cmd.PersistentFlags(), "output"))

	return cmd
}

func runExportCmd(opts ExportCmdOptions) error {
	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Tracef("finding files in API dir: %s", vppInput.ApiDirectory)

	files, err := vppapi.FindFiles(vppInput.ApiDirectory)
	if err != nil {
		return err
	}

	logrus.Debugf("found %d files in API dir", len(files))

	if err := os.Mkdir(opts.Output, 0750); err != nil {
		return err
	}

	logrus.Tracef("exporting files into output directory: %s", opts.Output)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		filerel, err := filepath.Rel(vppInput.ApiDirectory, f)
		if err != nil {
			logrus.Debugf("filepath.Rel error: %v", err)
			filerel = strings.TrimPrefix(f, vppInput.ApiDirectory)
		}
		filename := filepath.Join(opts.Output, filerel)
		dir := filepath.Dir(filename)

		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}

		if err := os.WriteFile(filename, data, 0660); err != nil {
			return err
		}

		logrus.Tracef("file %s exported (%d bytes) to %s", filerel, len(data), dir)
	}

	logrus.Debugf("exported %d files to %s", len(files), opts.Output)

	return nil
}
