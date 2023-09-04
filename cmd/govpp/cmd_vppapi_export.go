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

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

const exampleVppApiExportCmd = `
  <cyan># Export VPP API files to directory</>
  govpp vppapi export [INPUT] --output vppapi
  govpp vppapi export https://github.com/FDio/vpp --output vppapi

  <cyan># Export to archive</>
  govpp vppapi export [INPUT] --output vppapi.tar.gz
`

type VppApiExportCmdOptions struct {
	*VppApiCmdOptions

	Output string
	Targz  bool
	Flat   bool // TODO: use this
}

func newVppApiExportCmd(cli Cli, vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiExportCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:     "export [INPUT] --output OUTPUT [--targz] [--flat]",
		Short:   "Export VPP API files",
		Long:    "Export VPP API files from an input location to an output location.",
		Example: color.Sprint(exampleVppApiExportCmd),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			// auto-detect archive from output name
			if !cmd.Flags().Changed("targz") && strings.HasSuffix(opts.Output, ".tar.gz") {
				opts.Targz = true
			}
			return runExportCmd(opts)
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.Flat, "flat", false, "Export using flat structure")
	cmd.PersistentFlags().BoolVar(&opts.Targz, "targz", false, "Export to gzipped tarball archive")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output directory for the exported files")
	must(cobra.MarkFlagRequired(cmd.PersistentFlags(), "output"))

	return cmd
}

func runExportCmd(opts VppApiExportCmdOptions) error {
	vppInput, err := resolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	// collect files from input
	logrus.Tracef("preparing export from API dir: %s", vppInput.ApiDirectory)

	files, err := vppapi.FindFiles(vppInput.ApiDirectory)
	if err != nil {
		return err
	}

	logrus.Debugf("prepared %d API files for export from API dir: %s", len(files), vppInput.ApiDirectory)

	if opts.Targz {
		temp, err := os.MkdirTemp("", "govpp-vppapi-export-*")
		if err != nil {
			return fmt.Errorf("creating temp dir failed: %w", err)
		}

		tmpApiDir := filepath.Join(temp, "vppapi")

		err = exportVppInputToDir(tmpApiDir, files, vppInput)
		if err != nil {
			return fmt.Errorf("exporting files to temp dir failed: %w", err)
		}

		files = append(files, exportVPPVersionFile)

		err = copyFilesToTarGz(opts.Output, files, tmpApiDir)
		if err != nil {
			return fmt.Errorf("exporting to gzipped tarball failed: %w", err)
		}

		logrus.Debugf("exported %d files to archive: %s", len(files), opts.Output)
	} else {
		err = exportVppInputToDir(opts.Output, files, vppInput)
		if err != nil {
			return fmt.Errorf("exporting failed to dir failed: %w", err)
		}

		logrus.Debugf("exported %d files to: %s", len(files), opts.Output)
	}

	return nil
}

const exportVPPVersionFile = "VPP_VERSION"

func exportVppInputToDir(outputDir string, files []string, vppInput *vppapi.VppInput) error {
	logrus.Tracef("exporting %d files into directory: %s", len(files), outputDir)

	apiDir := vppInput.ApiDirectory

	// create the output directory
	if err := os.Mkdir(outputDir, 0750); err != nil {
		return fmt.Errorf("creating target dir failed: %w", err)
	}

	if err := exportFilesToDir(outputDir, files, apiDir); err != nil {
		return err
	}

	// write VPP version to file
	filename := filepath.Join(outputDir, exportVPPVersionFile)
	data := []byte(vppInput.Schema.Version)
	if err := os.WriteFile(filename, data, 0660); err != nil {
		return fmt.Errorf("writing VPP version file failed: %w", err)
	}

	return nil
}

func exportFilesToDir(outputDir string, files []string, apiDir string) error {
	logrus.Tracef("exporting %d files to output dir: %s", len(files), outputDir)

	// export files to directory
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(apiDir, file))
		if err != nil {
			return err
		}
		filename := filepath.Join(outputDir, file)
		if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
			return err
		}
		if err := os.WriteFile(filename, data, 0660); err != nil {
			return err
		}
		logrus.Tracef("- exported file %s (%d bytes) to %s", file, len(data), filename)
	}

	return nil
}

func copyFilesToTarGz(outputFile string, files []string, baseDir string) error {
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

	logrus.Tracef("copying %d files into archive: %s", len(files), outputFile)

	// copy files to archive
	for _, file := range files {
		// open the file for reading
		fr, err := os.Open(filepath.Join(baseDir, file))
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
		header.Name = strings.TrimPrefix(file, baseDir)

		logrus.Tracef("- copying file %s to: %s", file, header.Name)

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
