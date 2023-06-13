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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

// TODO:
//  - add option for exporting flat structure
//  - embed VPP version into export somehow

type VppApiExportCmdOptions struct {
	*VppApiCmdOptions

	Output string
	Targz  bool
}

func newVppApiExportCmd(cli Cli, vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiExportCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:   "export [INPUT] --output OUTPUT [--targz]",
		Short: "Export VPP API",
		Long:  "Export VPP API files from an input location to an output location.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			// auto-detect .tar.gz
			if !cmd.Flags().Changed("targz") && strings.HasSuffix(opts.Output, ".tar.gz") {
				opts.Targz = true
			}
			return runExportCmd(opts)
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.Targz, "targz", false, "Export to gzipped tarball")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output directory for the exported files")
	must(cobra.MarkFlagRequired(cmd.PersistentFlags(), "output"))

	return cmd
}

func runExportCmd(opts VppApiExportCmdOptions) error {
	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	// collect files from input
	logrus.Tracef("collecting files for export in API dir: %s", vppInput.ApiDirectory)

	files, err := vppapi.FindFiles(vppInput.ApiDirectory)
	if err != nil {
		return err
	}

	logrus.Debugf("exporting %d files", len(files))

	if opts.Targz {
		temp, err := os.MkdirTemp("", "govpp-vppapi-export-*")
		if err != nil {
			return fmt.Errorf("creating temp dir failed: %w", err)
		}
		tmpDir := filepath.Join(temp, "vppapi")
		err = exportFilesToDir(tmpDir, files, vppInput)
		if err != nil {
			return fmt.Errorf("exporting to directory failed: %w", err)
		}
		var files2 = []string{filepath.Join(tmpDir, "VPP_VERSION")}
		for _, f := range files {
			filerel, err := filepath.Rel(vppInput.ApiDirectory, f)
			if err != nil {
				filerel = strings.TrimPrefix(f, vppInput.ApiDirectory)
			}
			filename := filepath.Join(tmpDir, filerel)
			files2 = append(files2, filename)
		}
		err = exportFilesToTarGz(opts.Output, files2, tmpDir)
		if err != nil {
			return fmt.Errorf("exporting to gzipped tarball failed: %w", err)
		}
	} else {
		err = exportFilesToDir(opts.Output, files, vppInput)
		if err != nil {
			return fmt.Errorf("exporting failed: %w", err)
		}
	}

	logrus.Debugf("exported %d files to %s", len(files), opts.Output)

	return nil
}

func exportFilesToTarGz(outputFile string, files []string, baseDir string) error {
	logrus.Tracef("exporting %d files into tarball archive: %s", len(files), outputFile)

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
		//header.Name = filepath.Join(filepath.Base(baseDir), strings.TrimPrefix(file, apiDir))
		header.Name = strings.TrimPrefix(file, baseDir)

		logrus.Tracef("- exporting file: %q to: %s", file, header.Name)

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

func exportFilesToDir(outputDir string, files []string, vppInput *vppapi.VppInput) error {
	logrus.Tracef("exporting files into directory: %s", outputDir)

	apiDir := vppInput.ApiDirectory

	// create the output directory for export
	if err := os.Mkdir(outputDir, 0750); err != nil {
		return err
	}

	// export files to directory
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

	// write version
	filename := filepath.Join(outputDir, "VPP_VERSION")
	data := []byte(vppInput.Schema.Version)
	if err := os.WriteFile(filename, data, 0660); err != nil {
		return err
	}

	return nil
}
