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
	"sort"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

func SortFileObjectsByName(file *vppapi.File) {
	sort.SliceStable(file.Imports, func(i, j int) bool {
		return file.Imports[i] < file.Imports[j]
	})
	sort.SliceStable(file.EnumTypes, func(i, j int) bool {
		return file.EnumTypes[i].Name < file.EnumTypes[j].Name
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

func importedFiles(files []*vppapi.File, file *vppapi.File) []*vppapi.File {
	var list []*vppapi.File
	byName := func(s string) *vppapi.File {
		for _, f := range files {
			if f.Name == s {
				return f
			}
		}
		return nil
	}
	imported := map[string]struct{}{}
	for _, imp := range file.Imports {
		imp = normalizeImport(imp)
		impFile := byName(imp)
		if impFile == nil {
			log.Fatalf("file %q not found", imp)
		}
		for _, nest := range importedFiles(files, impFile) {
			if _, ok := imported[nest.Name]; !ok {
				list = append(list, nest)
				imported[nest.Name] = struct{}{}
			}
		}
		if _, ok := imported[impFile.Name]; !ok {
			list = append(list, impFile)
			imported[impFile.Name] = struct{}{}
		}
	}
	return list
}

func SortFilesByImports(apifiles []*vppapi.File) {
	dependsOn := func(file *vppapi.File, dep string) bool {
		for _, imp := range importedFiles(apifiles, file) {
			if imp.Name == dep {
				return true
			}
		}
		return false
	}
	sort.Slice(apifiles, func(i, j int) bool {
		a := apifiles[i]
		b := apifiles[j]
		if dependsOn(a, b.Name) {
			return false
		}
		if dependsOn(b, a.Name) {
			return true
		}
		return len(b.Imports) > len(a.Imports)
	})
}

func ListImportedTypes(apifiles []*vppapi.File, file *vppapi.File) []string {
	var importedTypes []string
	typeFiles := importedFiles(apifiles, file)
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

func RemoveImportedTypes(apifiles []*vppapi.File, apifile *vppapi.File) {
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
	apifile.AliasTypes = aliases
	apifile.StructTypes = structs
	apifile.UnionTypes = unions
}
