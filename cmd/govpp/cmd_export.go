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
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

// TODO:
// 	- exporting to archive (.tar.gz)
//  - add option for exporting flat structure
//  - embed VPP version into export somehow

type ExportCmdOptions struct {
	Input  string
	Output string
}

func newExportCmd() *cobra.Command {
	var (
		opts = ExportCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:   "export [INPUT] -o OUTPUT",
		Short: "Export VPP API files",
		Long:  "Export VPP API files from an input location to an output location.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
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

	// collect files from input
	logrus.Tracef("searching for files in API dir: %s", vppInput.ApiDirectory)

	files, err := vppapi.FindFiles(vppInput.ApiDirectory)
	if err != nil {
		return err
	}

	logrus.Debugf("found %d files in API dir %s", len(files), vppInput.ApiDirectory)

	if strings.HasSuffix(opts.Output, ".tar.gz") {
		err = exportFilesToTarGz(opts.Output, files, vppInput.ApiDirectory)
	} else {
		if err := os.Mkdir(opts.Output, 0750); err != nil {
			return err
		}

		err = exportFilesToDir(opts.Output, files, vppInput.ApiDirectory)
	}

	logrus.Debugf("exported %d files to %s", len(files), opts.Output)

	return nil
}

func exportFilesToTarGz(outputFile string, files []string, apiDir string) error {
	// create the output file for writing
	fw, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fw.Close()

	// create a new gzip writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// create a new tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		// open the file for reading
		fr, err := os.Open(file)
		if err != nil {
			return err
		}

		fi, err := fr.Stat()
		if err != nil {
			fr.Close()
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			fr.Close()
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = filepath.Join(filepath.Base(apiDir), strings.TrimPrefix(file, apiDir))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			fr.Close()
			return err
		}

		// copy file data to tar writer
		if _, err := io.Copy(tw, fr); err != nil {
			fr.Close()
			return err
		}

		fr.Close()
	}

	return nil
}

func exportFilesToDir(outputDir string, files []string, apiDir string) error {
	// export files to output
	logrus.Tracef("exporting files into output directory: %s", outputDir)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		filerel, err := filepath.Rel(apiDir, f)
		if err != nil {
			logrus.Debugf("filepath.Rel error: %v", err)
			filerel = strings.TrimPrefix(f, apiDir)
		}
		filename := filepath.Join(outputDir, filerel)
		dir := filepath.Dir(filename)

		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}

		if err := os.WriteFile(filename, data, 0660); err != nil {
			return err
		}

		logrus.Tracef("file %s exported (%d bytes) to %s", filerel, len(data), dir)
	}
	return nil
}
