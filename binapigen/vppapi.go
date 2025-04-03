//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"log"
	"path"
	"sort"
	"strings"

	"go.fd.io/govpp/binapigen/vppapi"
)

// SortFileObjectsByName sorts all objects of file by their name.
func SortFileObjectsByName(file *vppapi.File) {
	sort.SliceStable(file.Imports, func(i, j int) bool {
		return file.Imports[i] < file.Imports[j]
	})
	sort.SliceStable(file.EnumTypes, func(i, j int) bool {
		return file.EnumTypes[i].Name < file.EnumTypes[j].Name
	})
	sort.SliceStable(file.EnumflagTypes, func(i, j int) bool {
		return file.EnumflagTypes[i].Name < file.EnumflagTypes[j].Name
	})
	sort.Slice(file.AliasTypes, func(i, j int) bool {
		return file.AliasTypes[i].Name < file.AliasTypes[j].Name
	})
	sort.SliceStable(file.StructTypes, func(i, j int) bool {
		return file.StructTypes[i].Name < file.StructTypes[j].Name
	})
	sort.SliceStable(file.UnionTypes, func(i, j int) bool {
		return file.UnionTypes[i].Name < file.UnionTypes[j].Name
	})
	sort.SliceStable(file.Messages, func(i, j int) bool {
		return file.Messages[i].Name < file.Messages[j].Name
	})
	if file.Service != nil {
		sort.Slice(file.Service.RPCs, func(i, j int) bool {
			return file.Service.RPCs[i].Request < file.Service.RPCs[j].Request
		})
	}
}

func ListImportedFiles(files []vppapi.File, file *vppapi.File) []vppapi.File {
	var list []vppapi.File
	byName := func(s string) *vppapi.File {
		for _, f := range files {
			file := f
			if f.Name == s {
				return &file
			}
		}
		return nil
	}
	imported := map[string]struct{}{}
	for _, imp := range file.Imports {
		imp = normalizeImport(imp)
		impFile := byName(imp)
		if impFile == nil {
			log.Fatalf("imported file %q not found", imp)
		}
		for _, nest := range ListImportedFiles(files, impFile) {
			if _, ok := imported[nest.Name]; !ok {
				list = append(list, nest)
				imported[nest.Name] = struct{}{}
			}
		}
		if _, ok := imported[impFile.Name]; !ok {
			list = append(list, *impFile)
			imported[impFile.Name] = struct{}{}
		}
	}
	return list
}

// normalizeImport returns the last path element of the import, with all dotted suffixes removed.
func normalizeImport(imp string) string {
	imp = path.Base(imp)
	if idx := strings.Index(imp, "."); idx >= 0 {
		imp = imp[:idx]
	}
	return imp
}

// SortFilesByName sorts list of files by their name.
func SortFilesByName(apifiles []vppapi.File) {
	sort.SliceStable(apifiles, func(i, j int) bool {
		a := apifiles[i]
		b := apifiles[j]
		return a.Name < b.Name
	})
}

// SortFilesByImports sorts list of files by their imports.
func SortFilesByImports(apifiles []vppapi.File) {
	dependsOn := func(file *vppapi.File, dep string) bool {
		for _, imp := range ListImportedFiles(apifiles, file) {
			if imp.Name == dep {
				return true
			}
		}
		return false
	}
	sort.SliceStable(apifiles, func(i, j int) bool {
		a := apifiles[i]
		b := apifiles[j]
		if dependsOn(&a, b.Name) {
			return false
		}
		if dependsOn(&b, a.Name) {
			return true
		}
		return len(b.Imports) > len(a.Imports)
	})
}

// ListImportedTypes returns list of names for imported types.
func ListImportedTypes(apifiles []vppapi.File, file *vppapi.File) []string {
	var importedTypes []string
	typeFiles := ListImportedFiles(apifiles, file)
	for _, t := range file.StructTypes {
		var imported bool
		for _, imp := range typeFiles {
			for _, at := range imp.StructTypes {
				if at.Name != t.Name {
					continue
				}
				importedTypes = append(importedTypes, t.Name)
				imported = true
				break
			}
			if imported {
				break
			}
		}
	}
	for _, t := range file.AliasTypes {
		var imported bool
		for _, imp := range typeFiles {
			for _, at := range imp.AliasTypes {
				if at.Name != t.Name {
					continue
				}
				importedTypes = append(importedTypes, t.Name)
				imported = true
				break
			}
			if imported {
				break
			}
		}
	}
	for _, t := range file.EnumTypes {
		var imported bool
		for _, imp := range typeFiles {
			for _, at := range imp.EnumTypes {
				if at.Name != t.Name {
					continue
				}
				importedTypes = append(importedTypes, t.Name)
				imported = true
				break
			}
			if imported {
				break
			}
		}
	}
	for _, t := range file.EnumflagTypes {
		var imported bool
		for _, imp := range typeFiles {
			for _, at := range imp.EnumflagTypes {
				if at.Name != t.Name {
					continue
				}
				importedTypes = append(importedTypes, t.Name)
				imported = true
				break
			}
			if imported {
				break
			}
		}
	}
	for _, t := range file.UnionTypes {
		var imported bool
		for _, imp := range typeFiles {
			for _, at := range imp.UnionTypes {
				if at.Name != t.Name {
					continue
				}
				importedTypes = append(importedTypes, t.Name)
				imported = true
				break
			}
			if imported {
				break
			}
		}
	}
	return importedTypes
}

// RemoveImportedTypes removes imported types from file.
func RemoveImportedTypes(apifiles []vppapi.File, apifile *vppapi.File) {
	importedTypes := ListImportedTypes(apifiles, apifile)
	isImportedType := func(s string) bool {
		for _, t := range importedTypes {
			if t == s {
				return true
			}
		}
		return false
	}
	var enums []vppapi.EnumType
	for _, enumType := range apifile.EnumTypes {
		if !isImportedType(enumType.Name) {
			enums = append(enums, enumType)
		}
	}
	var enumflags []vppapi.EnumType
	for _, enumflagType := range apifile.EnumflagTypes {
		if !isImportedType(enumflagType.Name) {
			enumflags = append(enumflags, enumflagType)
		}
	}
	var aliases []vppapi.AliasType
	for _, aliasType := range apifile.AliasTypes {
		if !isImportedType(aliasType.Name) {
			aliases = append(aliases, aliasType)
		}
	}
	var structs []vppapi.StructType
	for _, structType := range apifile.StructTypes {
		if !isImportedType(structType.Name) {
			structs = append(structs, structType)
		}
	}
	var unions []vppapi.UnionType
	for _, unionType := range apifile.UnionTypes {
		if !isImportedType(unionType.Name) {
			unions = append(unions, unionType)
		}
	}
	apifile.EnumTypes = enums
	apifile.EnumflagTypes = enumflags
	apifile.AliasTypes = aliases
	apifile.StructTypes = structs
	apifile.UnionTypes = unions
}

// CleanMessageComment processes a comment string from VPP API message and
// returns a modified version with the following changes:
// - trim comment syntax ("/**", "*/")
// - remove special syntax ("\brief") parts
// - replace all occurrences of "@param" with a dash ("-").
func CleanMessageComment(comment string) string {
	// trim comment syntax
	comment = strings.TrimPrefix(comment, "/**")
	comment = strings.TrimSuffix(comment, " */")
	comment = strings.TrimSuffix(comment, "*/")

	// remove \\brief from the comment
	comment = strings.ReplaceAll(comment, `\\brief`, "")
	comment = strings.ReplaceAll(comment, `\brief`, "")

	// replace @param with a dash (-)
	comment = strings.ReplaceAll(comment, "@param", "-")

	return strings.TrimSpace(comment)
}

// StripMessageCommentFields processes a comment string from VPP API message and
// returns a modified version where a set of fields are omitted.
func StripMessageCommentFields(comment string, fields ...string) string {
	lines := strings.Split(comment, "\n")
	result := ""
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		add := true
		for _, field := range fields {
			if strings.Contains(line, " - "+field) {
				add = false
				break
			}
		}
		if add {
			result += line + "\n"
		}
	}
	return strings.TrimSuffix(result, "\n")
}

func normalizeCRC(crc string) string {
	return strings.TrimPrefix(crc, "0x")
}
