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
	"sort"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

// DifferenceType represents the type of difference found in the comparison.
type DifferenceType string

const (
	VersionDifference     DifferenceType = "Version"
	TotalFilesDifference  DifferenceType = "FilesCount"
	FileAddedDifference   DifferenceType = "FileAdded"
	FileRemovedDifference DifferenceType = "FileRemoved"

	FileMovedDifference           DifferenceType = "FileMoved"
	FileVersionDifference         DifferenceType = "FileVersion"
	FileCrcDifference             DifferenceType = "FileCRC"
	FileContentsChangedDifference DifferenceType = "FileContentsChanged"
	MessageAddedDifference        DifferenceType = "MessageAdded"
	MessageRemovedDifference      DifferenceType = "MessageRemoved"

	MessageCrcDifference       DifferenceType = "MessageCRC"
	MsgOptionChangedDifference DifferenceType = "MsgOptionChanged"
	MsgOptionAddedDifference   DifferenceType = "MsgOptionAdded"
	MsgOptionRemovedDifference DifferenceType = "MsgOptionRemoved"
	MessageCommentDifference   DifferenceType = "MessageComment"
)

var defaultDifferenceTypes = []DifferenceType{
	VersionDifference,
	TotalFilesDifference,
	FileAddedDifference,
	FileRemovedDifference,

	FileMovedDifference,
	FileVersionDifference,
	FileCrcDifference,
	FileContentsChangedDifference,
	MessageAddedDifference,
	MessageRemovedDifference,

	MessageCrcDifference,
	MsgOptionChangedDifference,
	MsgOptionAddedDifference,
	MsgOptionRemovedDifference,
}

// Difference represents a specific difference found between two schemas.
type Difference struct {
	Type           DifferenceType // Type is a type of the difference
	File           string         // File is a file name in which the difference was found in (or nil in case of schema-level differences)
	Description    string         // Description describes the difference
	Value1, Value2 any            // Value1 & Value2 contain the values that are being compared (or nil in case of add/remove-kind differences)
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
//
// Comparison process:
//
// 1. Schema-level properties
//   - schema version
//   - number of files
//   - added/removed files
//
// 2. File-level properties
//   - file version/CRC
//   - file path
//   - file options
//   - number of messages/types
//   - added/removed messages/types
//
// 3. Object-level properties
//   - message CRC
//   - message comment
//   - added/removed/changed message fields
//   - added/removed/changed message options
func CompareSchemas(schema1, schema2 *vppapi.Schema) []Difference {
	var differences []Difference

	logrus.Tracef("comparing schemas:\n\tSCHEMA 1: %v (%v files)\n\tSCHEMA 2: %v (%d files)\n", schema1.Version, len(schema1.Files), schema2.Version, len(schema2.Files))

	// compare VPP version
	if schema1.Version != schema2.Version {
		differences = append(differences, Difference{
			Type: VersionDifference,
			Description: color.Sprintf("Schema version is different: %s vs %s",
				clrDiffVersion.Sprint(schema1.Version), clrDiffVersion.Sprint(schema2.Version)),
			Value1: schema1.Version,
			Value2: schema2.Version,
		})
	}
	// compare file count
	if len(schema1.Files) != len(schema2.Files) {
		differences = append(differences, Difference{
			Type: TotalFilesDifference,
			Description: color.Sprintf("Total file count %s from %v to %v",
				clrWhite.Sprint(numberChangeString(len(schema1.Files), len(schema2.Files))),
				clrDiffNumber.Sprint(len(schema1.Files)), clrDiffNumber.Sprint(len(schema2.Files))),
			Value1: len(schema1.Files),
			Value2: len(schema2.Files),
		})
	}

	fileDiffs := compareSchemaFiles(schema1.Files, schema2.Files)
	logrus.Debugf("found %d differences between all schema files", len(fileDiffs))

	differences = append(differences, fileDiffs...)

	return differences
}

func compareSchemaFiles(files1 []vppapi.File, files2 []vppapi.File) []Difference {
	var differences []Difference

	// prepare files for comparison
	var fileList1 []string
	fileMap1 := make(map[string]vppapi.File)
	for _, file := range files1 {
		fileList1 = append(fileList1, file.Name)
		fileMap1[file.Name] = file
	}
	var fileList2 []string
	fileMap2 := make(map[string]vppapi.File)
	for _, file := range files2 {
		fileList2 = append(fileList2, file.Name)
		fileMap2[file.Name] = file
	}
	sort.Strings(fileList1)
	sort.Strings(fileList2)

	var fileCompare []string

	// removed files
	for _, fileName := range fileList1 {
		file1 := fileMap1[fileName]
		if _, ok := fileMap2[fileName]; ok {
			fileCompare = append(fileCompare, fileName)
		} else {
			differences = append(differences, Difference{
				Type:        FileRemovedDifference,
				Description: color.Sprintf("File removed: %s", clrDiffFile.Sprint(fileName)),
				Value1:      file1,
				Value2:      nil,
			})
		}
	}
	// added files
	for _, fileName := range fileList2 {
		file2 := fileMap2[fileName]
		if _, ok := fileMap1[fileName]; !ok {
			differences = append(differences, Difference{
				Type:        FileAddedDifference,
				Description: color.Sprintf("File added: %s", clrDiffFile.Sprint(fileName)),
				Value1:      nil,
				Value2:      file2,
			})
		}
	}
	sort.Strings(fileCompare)

	// changed files
	for _, fileName := range fileCompare {
		file1 := fileMap1[fileName]
		file2 := fileMap2[fileName]
		fileDiffs := compareFiles(file1, file2)
		if len(fileDiffs) > 0 {
			logrus.Tracef("found %2d differences between files: %s", len(fileDiffs), fileName)
		}
		for _, fileDiff := range fileDiffs {
			fileDiff.File = fileName
			differences = append(differences, fileDiff)
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
				clrDiffVersion.Sprint(fileVer1), clrDiffVersion.Sprint(fileVer2)),
			Value1: fileVer1,
			Value2: fileVer2,
		})
	}
	if file1.CRC != file2.CRC {
		differences = append(differences, Difference{
			Type: FileCrcDifference,
			Description: color.Sprintf("File CRC changed from %s to %s",
				clrDiffVersion.Sprint(file1.CRC), clrDiffVersion.Sprint(file2.CRC)),
			Value1: file1.CRC,
			Value2: file2.CRC,
		})
	}
	// TODO: compare other file options

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
				if msgDiff.Value1 == nil {
					msgDiff.Value1 = msg1
				}
				if msgDiff.Value2 == nil {
					msgDiff.Value2 = msg2
				}
				differences = append(differences, msgDiff)
			}
		} else {
			differences = append(differences, Difference{
				Type:        MessageRemovedDifference,
				Description: color.Sprintf("Message removed: %s", clrDiffMessage.Sprint(msgName)),
				Value1:      msg1,
				Value2:      nil,
			})
		}
	}
	// added messages
	for msgName, msg := range msgMap2 {
		if _, ok := msgMap1[msgName]; !ok {
			differences = append(differences, Difference{
				Type:        MessageAddedDifference,
				Description: color.Sprintf("Message added: %s", clrDiffMessage.Sprint(msgName)),
				Value1:      nil,
				Value2:      msg,
			})
		}
	}

	// TODO: compare added/removed types

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
				clrDiffMessage.Sprint(msg1.Name), clrDiffVersion.Sprint(msg1.CRC), clrDiffVersion.Sprint(msg2.CRC)),
			Value1: msg1.CRC,
			Value2: msg2.CRC,
		})
	}

	// Compare message comments
	if msg1.Comment != msg2.Comment {
		desc := color.Sprintf("Message %s comment ", clrDiffMessage.Sprint(msg1.Name))
		if msg1.Comment == "" {
			desc += "added"
		} else if msg2.Comment == "" {
			desc += "removed"
		} else {
			desc += "changed"
		}
		differences = append(differences, Difference{
			Type:        MessageCommentDifference,
			Description: desc,
			Value1:      msg1.Comment,
			Value2:      msg2.Comment,
		})
	}

	// removed options
	for option, val1 := range msg1.Options {
		if val2, ok := msg2.Options[option]; ok {
			if val1 != val2 {
				differences = append(differences, Difference{
					Type: MsgOptionChangedDifference,
					Description: color.Sprintf("Message %s changed option %s from %q to %q",
						clrDiffMessage.Sprint(msg1.Name), clrDiffOption.Sprint(option), clrDiffOption.Sprint(val1), clrDiffOption.Sprint(val2)),
					Value1: keyValString(option, val1),
					Value2: keyValString(option, val2),
				})
			}
		} else {
			differences = append(differences, Difference{
				Type: MsgOptionRemovedDifference,
				Description: color.Sprintf("Message %s removed option: %s",
					clrDiffMessage.Sprint(msg1.Name), clrDiffOption.Sprint(keyValString(option, val1))),
				Value1: keyValString(option, val1),
				Value2: "",
			})
		}
	}
	// added options
	for option, val2 := range msg2.Options {
		if _, ok := msg1.Options[option]; !ok {
			differences = append(differences, Difference{
				Type: MsgOptionAddedDifference,
				Description: color.Sprintf("Message %s added option: %s",
					clrDiffMessage.Sprint(msg2.Name), clrDiffOption.Sprint(keyValString(option, val2))),
				Value1: "",
				Value2: keyValString(option, val2),
			})
		}
	}

	return differences
}

func numberOfContentChangedDifference(typ string, c1, c2 int) Difference {
	return Difference{
		Type: FileContentsChangedDifference,
		Description: color.Sprintf("Number of %s has %s from %v to %v",
			clrWhite.Sprint(typ), clrWhite.Sprint(numberChangeString(c1, c2)), clrDiffNumber.Sprint(c1), clrDiffNumber.Sprint(c2)),
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
	return color.Sprintf("%s=%q", k, v)
}
