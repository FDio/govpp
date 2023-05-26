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
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

// DifferenceType represents the type of difference found in the comparison.
type DifferenceType string

const (
	VersionDifference    DifferenceType = "Version"
	TotalFilesDifference DifferenceType = "FilesCount"

	FileAddedDifference           DifferenceType = "FileAdded"
	FileRemovedDifference         DifferenceType = "FileRemoved"
	FileMovedDifference           DifferenceType = "FileMoved"
	FileVersionDifference         DifferenceType = "FileVersion"
	FileCrcDifference             DifferenceType = "FileCRC"
	FileContentsChangedDifference DifferenceType = "FileContentsChanged"

	MessageAddedDifference   DifferenceType = "MessageAdded"
	MessageRemovedDifference DifferenceType = "MessageRemoved"
	MessageCrcDifference     DifferenceType = "MessageCRC"

	MsgOptionChangedDifference DifferenceType = "MsgOptionChanged"
	MsgOptionAddedDifference   DifferenceType = "MsgOptionAdded"
	MsgOptionRemovedDifference DifferenceType = "MsgOptionRemoved"
)

// Difference represents a difference between two API schemas.
type Difference struct {
	Type           DifferenceType
	File           string
	Description    string
	Value1, Value2 any
}

func (d Difference) String() string {
	var b strings.Builder
	color.Fprintf(&b, "[%v] ", clrWhite.Sprint(d.Type))
	if d.File != "" {
		color.Fprintf(&b, "%s: ", clrDiffFile.Sprint(d.File))
	}
	color.Fprintf(&b, "%v", d.Description)
	return b.String()
}

// CompareSchemas compares two API schemas and returns a list of differences.
func CompareSchemas(schema1, schema2 *vppapi.Schema) []Difference {
	var differences []Difference

	// compare VPP version
	if schema1.Version != schema2.Version {
		differences = append(differences, Difference{
			Type: VersionDifference,
			Description: color.Sprintf("Schema version is different: %s vs %s",
				clrWhite.Sprint(schema1.Version), clrWhite.Sprint(schema2.Version)),
			Value1: schema1.Version,
			Value2: schema2.Version,
		})
	}
	// compare file count
	if len(schema1.Files) != len(schema2.Files) {
		differences = append(differences, Difference{
			Type: TotalFilesDifference,
			Description: color.Sprintf("Total file count %s from %v to %v",
				clrWhite.Sprint(numberChangeString(len(schema1.Files), len(schema2.Files))), clrWhite.Sprint(len(schema1.Files)), clrWhite.Sprint(len(schema2.Files))),
			Value1: len(schema1.Files),
			Value2: len(schema2.Files),
		})
	}

	// compare schema files
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
				Description: "File removed",
				Value1:      file1,
				Value2:      nil,
			})
		}
	}
	// added files
	for fileName, file2 := range fileMap2 {
		if _, ok := fileMap1[fileName]; !ok {
			differences = append(differences, Difference{
				Type:        FileAddedDifference,
				File:        fileName,
				Description: "File added",
				Value1:      nil,
				Value2:      file2,
			})
		}
	}
	return differences
}

// compareFiles compares two files and returns a list of differences.
func compareFiles(file1, file2 vppapi.File) []Difference {
	if file1.Name != file2.Name {
		logrus.Fatalf("comparing different files (%s vs. %s)", file1.Name, file2.Name)
	}

	var differences []Difference

	// Compare file properties
	if file1.Path != file2.Path {
		differences = append(differences, Difference{
			Type: FileMovedDifference,
			Description: color.Sprintf("File moved from %s to %s",
				clrWhite.Sprint(file1.Path), clrWhite.Sprint(file2.Path)),
			Value1: file1.Path,
			Value2: file2.Path,
		})
	}
	if fileVer1, fileVer2 := file1.Options[vppapi.OptFileVersion], file2.Options[vppapi.OptFileVersion]; fileVer1 != fileVer2 {
		differences = append(differences, Difference{
			Type: FileVersionDifference,
			Description: color.Sprintf("File version changed from %s to %s",
				clrWhite.Sprint(fileVer1), clrWhite.Sprint(fileVer2)),
			Value1: fileVer1,
			Value2: fileVer2,
		})
	}
	if file1.CRC != file2.CRC {
		differences = append(differences, Difference{
			Type: FileCrcDifference,
			Description: color.Sprintf("File CRC changed from %s to %s",
				clrWhite.Sprint(file1.CRC), clrWhite.Sprint(file2.CRC)),
			Value1: file1.CRC,
			Value2: file2.CRC,
		})
	}

	// Compare number of messages and types
	if len(file1.Messages) != len(file2.Messages) {
		differences = append(differences, numberOfContentChangedDifference("Messages", len(file1.Messages), len(file2.Messages)))
	}
	if len(file1.StructTypes) != len(file2.StructTypes) {
		differences = append(differences, numberOfContentChangedDifference("Types", len(file1.StructTypes), len(file2.StructTypes)))
	}
	if len(file1.UnionTypes) != len(file2.UnionTypes) {
		differences = append(differences, numberOfContentChangedDifference("Unions", len(file1.UnionTypes), len(file2.UnionTypes)))
	}
	if len(file1.AliasTypes) != len(file2.AliasTypes) {
		differences = append(differences, numberOfContentChangedDifference("Aliases", len(file1.AliasTypes), len(file2.AliasTypes)))
	}
	if len(file1.EnumTypes) != len(file2.EnumTypes) {
		differences = append(differences, numberOfContentChangedDifference("Enums", len(file1.EnumTypes), len(file2.EnumTypes)))
	}
	if len(file1.EnumflagTypes) != len(file2.EnumflagTypes) {
		differences = append(differences, numberOfContentChangedDifference("Enumflags", len(file1.EnumflagTypes), len(file2.EnumflagTypes)))
	}

	// Compare file messages
	msgMap1 := make(map[string]vppapi.Message)
	for _, msg := range file1.Messages {
		msgMap1[msg.Name] = msg
	}
	msgMap2 := make(map[string]vppapi.Message)
	for _, msg := range file2.Messages {
		msgMap2[msg.Name] = msg
	}
	// removed messages
	for msgName, msg1 := range msgMap1 {
		if msg2, ok := msgMap2[msgName]; ok {
			msgDiffs := compareMessages(msg1, msg2)
			for _, msgDiff := range msgDiffs {
				msgDiff.Value1 = msg1
				msgDiff.Value2 = msg2
				differences = append(differences, msgDiff)
			}
		} else {
			differences = append(differences, Difference{
				Type:        MessageRemovedDifference,
				Description: color.Sprintf("Message removed: %s", clrCyan.Sprint(msgName)),
				Value1:      msg1,
				Value2:      nil,
			})
		}
	}
	// added messages
	for msgName := range msgMap2 {
		if msg2, ok := msgMap1[msgName]; !ok {
			differences = append(differences, Difference{
				Type:        MessageAddedDifference,
				Description: color.Sprintf("Message added: %s", clrCyan.Sprint(msgName)),
				Value1:      nil,
				Value2:      msg2,
			})
		}
	}

	return differences
}

func compareMessages(msg1 vppapi.Message, msg2 vppapi.Message) []Difference {
	if msg1.Name != msg2.Name {
		logrus.Fatalf("comparing different messages (%s vs. %s)", msg1.Name, msg2.Name)
	}

	var differences []Difference

	// Compare message properties
	if msg1.CRC != msg2.CRC {
		differences = append(differences, Difference{
			Type: MessageCrcDifference,
			Description: color.Sprintf("Message %s changed CRC from %s to %s",
				clrCyan.Sprint(msg1.Name), msg1.CRC, msg2.CRC),
		})
	}

	// removed options
	for option, val1 := range msg1.Options {
		if val2, ok := msg2.Options[option]; ok {
			if val1 != val2 {
				differences = append(differences, Difference{
					Type: MsgOptionChangedDifference,
					Description: color.Sprintf("Message %s changed option %s from %q to %q",
						clrCyan.Sprint(msg1.Name), clrWhite.Sprint(option), clrWhite.Sprint(val1), clrWhite.Sprint(val2)),
				})
			}
		} else {
			differences = append(differences, Difference{
				Type: MsgOptionRemovedDifference,
				Description: color.Sprintf("Message %s removed option: %s",
					clrCyan.Sprint(msg1.Name), clrWhite.Sprint(option)),
				Value1: keyValString(option, val1),
				Value2: nil,
			})
		}
	}
	// added options
	for option, val := range msg2.Options {
		if _, ok := msg1.Options[option]; !ok {
			differences = append(differences, Difference{
				Type: MsgOptionAddedDifference,
				Description: color.Sprintf("Message %s added option: %s",
					clrCyan.Sprint(msg2.Name), clrWhite.Sprint(keyValString(option, val))),
				Value1: nil,
				Value2: keyValString(option, val),
			})
		}
	}

	return differences
}

func numberOfContentChangedDifference(typ string, c1, c2 int) Difference {
	return Difference{
		Type: FileContentsChangedDifference,
		Description: color.Sprintf("Number of %s has %s from %v to %v",
			clrWhite.Sprint(typ), clrWhite.Sprint(numberChangeString(c1, c2)), clrWhite.Sprint(c1), clrWhite.Sprint(c2)),
		Value1: c1,
		Value2: c2,
	}
}

func numberChangeString(n1, n2 int) string {
	switch {
	case n1 < n2:
		return "increased"
	case n1 > n2:
		return "decreased"
	default:
		return ""
	}
}

func keyValString(k, v string) string {
	if v == "" {
		return k
	}
	return color.Sprintf("%s=%s", k, v)
}
