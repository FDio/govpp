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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

type DiffCmdOptions struct {
	Input   string
	Against string
}

func newDiffCmd() *cobra.Command {
	var (
		opts = DiffCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:     "diff INPUT --against=AGAINST",
		Aliases: []string{"dif", "d", "cmp", "compare", "changes"},
		Short:   "b two schemas",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Input = args[0]
			return runDiffCmd(opts)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Against, "against", "", "Required. The schema to compare against.")

	return cmd
}

func runDiffCmd(opts DiffCmdOptions) error {
	// Input
	vppInput, err := vppapi.ResolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Tracef("VPP input:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppInput.ApiDirectory, vppInput.Schema.Version, len(vppInput.Schema.Files))

	// Against
	vppAgainst, err := vppapi.ResolveVppInput(opts.Against)
	if err != nil {
		return err
	}

	logrus.Tracef("VPP against:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppAgainst.ApiDirectory, vppAgainst.Schema.Version, len(vppAgainst.Schema.Files))

	schema1 := vppInput.Schema
	schema2 := vppAgainst.Schema

	logrus.Debugf("comparing schemas..")

	diffs := CompareSchemas(&schema1, &schema2)

	if len(diffs) == 0 {
		fmt.Printf("")
		return nil
	}

	fmt.Printf("Listing %d differences:\n", len(diffs))
	for _, d := range diffs {
		if d.File != "" {
			fmt.Printf(" - [%v] %s: %v\n", d.Type, d.File, d.Description)
		} else {
			fmt.Printf(" - [%v] %v\n", d.Type, d.Description)
		}
	}

	return nil
}

// DifferenceType represents the type of difference found in the comparison.
type DifferenceType string

const (
	VersionDifference        DifferenceType = "Version"
	TotalFilesDifference     DifferenceType = "FilesCount"
	FileAddedDifference      DifferenceType = "FileAdded"
	FileRemovedDifference    DifferenceType = "FileRemoved"
	FileMovedDifference      DifferenceType = "FileMoved"
	FileVersionDifference    DifferenceType = "FileVersion"
	ContentChangedDifference DifferenceType = "FileChanged"
)

// Difference represents a difference in the API schemas.
type Difference struct {
	Type        DifferenceType
	File        string
	Description string
}

// CompareSchemas compares two API schemas and returns a list of differences.
func CompareSchemas(schema1, schema2 *vppapi.Schema) []Difference {
	var differences []Difference

	// compare version
	if schema1.Version != schema2.Version {
		differences = append(differences, Difference{
			Type:        VersionDifference,
			Description: fmt.Sprintf("Versions are different: %s vs %s", schema1.Version, schema2.Version),
		})
	}
	// compare file count
	if len(schema1.Files) != len(schema2.Files) {
		differences = append(differences, Difference{
			Type:        TotalFilesDifference,
			Description: fmt.Sprintf("Total file count is different: %d vs %d", len(schema1.Files), len(schema2.Files)),
		})
	}

	// prepare file maps
	fileMap1 := make(map[string]vppapi.File)
	for _, file := range schema1.Files {
		fileMap1[file.Name] = file
	}
	fileMap2 := make(map[string]vppapi.File)
	for _, file := range schema2.Files {
		fileMap2[file.Name] = file
	}
	// removed files
	for fileName, file1 := range fileMap1 {
		if file2, ok := fileMap2[fileName]; ok {
			fileDiffs := compareFiles(file1, file2)
			for _, fileDiff := range fileDiffs {
				fileDiff.File = fileName
				differences = append(differences, fileDiff)
			}
		} else {
			differences = append(differences, Difference{
				Type:        FileRemovedDifference,
				File:        fileName,
				Description: fmt.Sprintf("File removed"),
			})
		}
	}
	// added files
	for fileName := range fileMap2 {
		if _, ok := fileMap1[fileName]; !ok {
			differences = append(differences, Difference{
				Type:        FileAddedDifference,
				File:        fileName,
				Description: fmt.Sprintf("File added"),
			})
		}
	}
	return differences
}

// compareFiles compares two files and returns a list of differences.
func compareFiles(file1, file2 vppapi.File) []Difference {
	var differences []Difference

	// Compare file properties
	if file1.Path != file2.Path {
		differences = append(differences, Difference{
			Type:        FileMovedDifference,
			Description: fmt.Sprintf("Path changed from %s to %s", file1.Path, file2.Path),
		})
	}
	if file1.Options[vppapi.OptFileVersion] != file2.Options[vppapi.OptFileVersion] {
		differences = append(differences, Difference{
			Type:        FileVersionDifference,
			Description: fmt.Sprintf("Version changed from %s to %s", file1.Options[vppapi.OptFileVersion], file2.Options[vppapi.OptFileVersion]),
		})
	}
	if file1.CRC != file2.CRC {
		differences = append(differences, Difference{
			Type:        FileVersionDifference,
			Description: fmt.Sprintf("File CRC changed from %s to %s", file1.CRC, file2.CRC),
		})
	}

	// Compare number of messages and types
	if len(file1.Messages) != len(file2.Messages) {
		differences = append(differences, fileChangedDifference("Messages", len(file1.Messages), len(file2.Messages)))
	}
	if len(file1.StructTypes) != len(file2.StructTypes) {
		differences = append(differences, fileChangedDifference("Types", len(file1.StructTypes), len(file2.StructTypes)))
	}
	if len(file1.UnionTypes) != len(file2.UnionTypes) {
		differences = append(differences, fileChangedDifference("Unions", len(file1.UnionTypes), len(file2.UnionTypes)))
	}
	if len(file1.AliasTypes) != len(file2.AliasTypes) {
		differences = append(differences, fileChangedDifference("Aliases", len(file1.AliasTypes), len(file2.AliasTypes)))
	}
	if len(file1.EnumTypes) != len(file2.EnumTypes) {
		differences = append(differences, fileChangedDifference("Enums", len(file1.EnumTypes), len(file2.EnumTypes)))
	}
	if len(file1.EnumflagTypes) != len(file2.EnumflagTypes) {
		differences = append(differences, fileChangedDifference("Enumflags", len(file1.EnumflagTypes), len(file2.EnumflagTypes)))
	}

	return differences
}

func fileChangedDifference(typ string, c1, c2 int) Difference {
	var change string
	if c1 < c2 {
		change = "increased"
	} else {
		change = "decreased"
	}
	return Difference{
		Type:        ContentChangedDifference,
		Description: fmt.Sprintf("Number of %s has %s from %d to %d", typ, change, c1, c2),
	}
}
