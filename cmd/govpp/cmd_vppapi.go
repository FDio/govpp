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
	"bytes"
	"fmt"
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
)

type ApiOptions struct {
	Input           string
	Format          string
	ShowContents    bool
	ShowMessages    bool
	ShowRPC         bool
	IncludeImported bool
	IncludeFields   bool
}

func newVppapiCmd() *cobra.Command {
	var (
		opts = ApiOptions{
			Input: vppapi.DefaultDir,
		}
	)
	cmd := &cobra.Command{
		Use:     "api [FILE, ...]",
		Aliases: []string{"a", "vppapi"},
		Short:   "Print VPP API contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApi(cmd.OutOrStdout(), opts, args)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", opts.Input, "Path to directory containing VPP API files")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "", "Output format (json, yaml, go-template..)")
	cmd.PersistentFlags().BoolVar(&opts.ShowContents, "show-contents", false, "Show contents of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowRPC, "show-rpc", false, "Show service RPCs of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowMessages, "show-messages", false, "Show messages for VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.IncludeImported, "include-imported", false, "Include imported types")
	cmd.PersistentFlags().BoolVar(&opts.IncludeFields, "include-fields", false, "Include message fields")

	return cmd
}

func runApi(out io.Writer, opts ApiOptions, args []string) error {
	vppInput, err := vppapi.ResolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Tracef("VPP input:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppInput.ApiDirectory, vppInput.VppVersion, len(vppInput.ApiFiles))

	logrus.Debugf("parsing VPP API directory: %v", opts.Input)

	// parse API directory
	allapifiles, err := vppapi.ParseDir(opts.Input)
	if err != nil {
		return fmt.Errorf("vppapi.ParseDir: %w", err)
	}

	logrus.Debugf("parsed %d files, normalizing data", len(allapifiles))

	// normalize data
	binapigen.SortFilesByImports(allapifiles)

	// filter files
	var apifiles []vppapi.File
	if len(args) > 0 {
		// select only specific files
		for _, arg := range args {
			var found bool
			for _, apifile := range allapifiles {
				if apifile.Name == arg {
					apifiles = append(apifiles, apifile)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("VPP API file %q not found", arg)
			}
		}
	} else {
		// select all files
		apifiles = allapifiles
	}

	logrus.Debugf("selected %d/%d files", len(apifiles), len(allapifiles))

	// omit imported types
	if !opts.IncludeImported {
		for _, apifile := range apifiles {
			binapigen.RemoveImportedTypes(allapifiles, &apifile)
		}
	}

	// format output
	if format := opts.Format; len(format) == 0 {
		if opts.ShowMessages {
			apimsgs := listVPPAPIMessages(apifiles)
			showVPPAPIMessages(out, apimsgs, opts.IncludeFields)
		} else if opts.ShowRPC {
			showVPPAPIRPCs(out, apifiles)
		} else if opts.ShowContents {
			showVPPAPIContents(out, apifiles)
		} else {
			showVppApiList(out, apifiles)
		}
	} else {
		if err := formatAsTemplate(out, format, apifiles); err != nil {
			return err
		}
	}

	return nil
}

type ApiFile struct {
	File        string
	Version     string
	Path        string
	NumImports  uint
	NumMessages uint
	NumTypes    uint
	NumRPCs     uint
}

func showVppApiList(out io.Writer, apifiles []vppapi.File) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "API", "Version", "CRC", "Path", "Imports", "Messages", "Types", "RPCs",
	})
	//table.SetAutoMergeCells(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_RIGHT, 0, 0, 0, 0, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
	})

	for i, apifile := range apifiles {

		index := i + 1
		//importedTypes := binapigen.ListImportedTypes(apifiles, apifile)
		typesCount := len(apifile.EnumTypes) + len(apifile.EnumflagTypes) + len(apifile.AliasTypes) + len(apifile.StructTypes) + len(apifile.UnionTypes)
		//importedFiles := listImportedFiles(apifiles, apifile)
		apiVersion := "0.0.0"
		var options []string
		for k, v := range apifile.Options {
			if k == binapigen.OptFileVersion {
				apiVersion = v
				continue
			}
			options = append(options, fmt.Sprintf("%s=%v", k, v))
		}
		var imports string
		if len(apifile.Imports) > 0 {
			var importList []string
			if len(apifile.Imports) > 0 {
				importList = append(importList, fmt.Sprintf("%d", len(apifile.Imports)))
			}
			imports = strings.Join(importList, ", ")
		} else {
			imports = "-"
		}
		apiPath := pathOfParentAndFile(apifile.Path)
		apiPath = path.Dir(apiPath)

		var msgs string
		if len(apifile.Messages) > 0 {
			msgs = fmt.Sprintf("%3d", len(apifile.Messages))
		} else {
			msgs = fmt.Sprintf("%3s", "-")
		}
		var types string
		if typesCount > 0 {
			types = fmt.Sprintf("%2d", typesCount)
		} else {
			types = fmt.Sprintf("%2s", "-")
		}
		var services string
		if apifile.Service != nil {
			services = fmt.Sprintf("%2d", len(apifile.Service.RPCs))
		} else {
			services = fmt.Sprintf("%2s", "-")
		}

		var row []string
		row = []string{
			fmt.Sprint(index), apifile.Name, apiVersion, apifile.CRC, apiPath, imports, msgs, types, services,
		}
		table.Append(row)
	}
	table.Render()
}

func showVppApiListOld(out io.Writer, apifiles []*vppapi.File) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 1, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "#\tAPI\tVERSION\tPATH\tIMPORTS\tMESSAGES\tTYPES\tRPCs\t\n")

	for i, apifile := range apifiles {
		index := i + 1
		typesCount := len(apifile.EnumTypes) + len(apifile.EnumflagTypes) + len(apifile.AliasTypes) + len(apifile.StructTypes) + len(apifile.UnionTypes)
		apiVersion := "0.0.0"
		var options []string
		for k, v := range apifile.Options {
			if k == binapigen.OptFileVersion {
				apiVersion = v
				continue
			}
			options = append(options, fmt.Sprintf("%s=%v", k, v))
		}
		if apifile.CRC != "" {
			apiVersion += fmt.Sprintf("-%s", strings.TrimPrefix(apifile.CRC, "0x"))
		}
		var imports string
		if len(apifile.Imports) > 0 {
			var importList []string
			if len(apifile.Imports) > 0 {
				importList = append(importList, fmt.Sprintf("%d", len(apifile.Imports)))
			}
			imports = strings.Join(importList, ", ")
		} else {
			imports = "-"
		}
		apiPath := pathOfParentAndFile(apifile.Path)
		apiPath = path.Dir(apiPath)

		var msgs string
		if len(apifile.Messages) > 0 {
			msgs = fmt.Sprintf("%3d", len(apifile.Messages))
		} else {
			msgs = fmt.Sprintf("%3s", "-")
		}
		var types string
		if typesCount > 0 {
			types = fmt.Sprintf("%2d", typesCount)
		} else {
			types = fmt.Sprintf("%2s", "-")
		}
		var services string
		if apifile.Service != nil {
			services = fmt.Sprintf("%2d", len(apifile.Service.RPCs))
		} else {
			services = fmt.Sprintf("%2s", "-")
		}
		fmt.Fprintf(w, "%v\t%s\t%s\t%s\t%s\t%v\t%s\t%v\t\n",
			index, apifile.Name, apiVersion, apiPath, imports, msgs, types, services)
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(out, buf.String())
}

func pathOfParentAndFile(path string) string {
	// Split the path into components separated by "/"
	components := strings.Split(path, "/")

	// If there are less than 2 components, the output is the same as the input
	if len(components) < 2 {
		return path
	}

	// Return the last two components joined by "/"
	return strings.Join(components[len(components)-2:], "/")
}

const maxCommentLengthInColumn = 80

func showVPPAPIContents(out io.Writer, apifiles []vppapi.File) {
	var buf bytes.Buffer

	for _, apifile := range apifiles {
		apiVersion := "0.0.0"
		var options []string
		for k, v := range apifile.Options {
			if k == binapigen.OptFileVersion {
				apiVersion = v
				continue
			}
			options = append(options, fmt.Sprintf("%s=%v", k, v))
		}
		if apifile.CRC != "" {
			apiVersion += fmt.Sprintf("-%s", strings.TrimPrefix(apifile.CRC, "0x"))
		}
		fmt.Fprintln(&buf, "--------------------------------------------------------------------------------")
		fmt.Fprintf(&buf, "# %v (%s)\n", apifile.Name, apiVersion)
		if len(options) > 0 {
			fmt.Fprintf(&buf, "# Options: %v\n", mapStr(apifile.Options))
		}
		fmt.Fprintln(&buf, "--------------------------------------------------------------------------------")
		fmt.Fprintln(&buf)

		// Messages
		if len(apifile.Messages) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "MESSAGE", "CRC", "FIELDS", "OPTIONS", "COMMENT"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, msg := range apifile.Messages {
				index := i + 1
				msgCrc := getMsgCrc(msg.CRC)
				msgFields := fmt.Sprintf("%v", len(msg.Fields))
				msgOptions := getMsgOptionsShort(msg.Options)
				msgComment := getMsgCommentSimple(msg.Comment)
				row := []string{strconv.Itoa(index), msg.Name, msgCrc, msgFields, msgOptions, msgComment}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))
		}

		// Structs
		if len(apifile.StructTypes) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "TYPE", "FIELDS"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, typ := range apifile.StructTypes {
				index := i + 1
				fields := fmt.Sprintf("%v", len(typ.Fields))
				row := []string{strconv.Itoa(index), typ.Name, fields}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))
		}

		// Unions
		if len(apifile.UnionTypes) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "UNION", "FIELDS"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, typ := range apifile.UnionTypes {
				index := i + 1
				fields := fmt.Sprintf("%v", len(typ.Fields))
				row := []string{strconv.Itoa(index), typ.Name, fields}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))
		}

		// Enums
		if len(apifile.EnumTypes) > 0 || len(apifile.EnumflagTypes) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "ENUM", "TYPE", "KIND", "ENTRIES"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, typ := range apifile.EnumTypes {
				index := i + 1
				typEntries := fmt.Sprintf("%v", len(typ.Entries))
				row := []string{strconv.Itoa(index), typ.Name, typ.Type, "enum", typEntries}
				table.Append(row)
			}
			for i, typ := range apifile.EnumflagTypes {
				index := i + 1
				typEntries := fmt.Sprintf("%v", len(typ.Entries))
				row := []string{strconv.Itoa(index), typ.Name, typ.Type, "flag", typEntries}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))
		}

		// Aliases

		if len(apifile.AliasTypes) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "ALIAS", "TYPE", "LENGTH"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, typ := range apifile.AliasTypes {
				index := i + 1
				row := []string{strconv.Itoa(index), typ.Name, typ.Type, strconv.Itoa(typ.Length)}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))

		}

	}

	fmt.Fprint(out, buf.String())
}

func getMsgCrc(crc string) string {
	return strings.TrimPrefix(crc, "0x")
}

const maxLengthMsgOptionValue = 30

func getMsgOptionsShort(options map[string]string) string {
	opts := make(map[string]string, len(options))
	for k, v := range options {
		if len(v) > maxLengthMsgOptionValue {
			opts[k] = v[:maxLengthMsgOptionValue] + "..."
		} else {
			opts[k] = v
		}
	}
	return fmt.Sprintf("%v", mapStrOrdered(opts))
}

func getMsgCommentSimple(comment string) string {
	if comment == "" {
		return "-"
	}

	// get first line
	commentParts := strings.Split(comment, "\n")
	if len(commentParts) > 1 && strings.TrimSpace(commentParts[0]) == "/*" {
		commentParts = commentParts[1:]
	}

	msgComment := strings.TrimSpace(commentParts[0])

	// strip useless parts
	msgComment = strings.TrimPrefix(msgComment, "*")
	msgComment = strings.TrimPrefix(msgComment, "/**")
	msgComment = strings.ReplaceAll(msgComment, "\\brief", "")
	msgComment = strings.ReplaceAll(msgComment, "@brief", "")
	msgComment = strings.TrimSpace(msgComment)

	// cut long comments
	if len(msgComment) > maxCommentLengthInColumn {
		msgComment = msgComment[:maxCommentLengthInColumn] + "..."
	}

	return msgComment
}

type ApiFileMessage struct {
	File *vppapi.File
	vppapi.Message
}

func listVPPAPIMessages(apifiles []vppapi.File) []ApiFileMessage {
	var msgs []ApiFileMessage
	for _, apifile := range apifiles {
		file := apifile
		for _, msg := range apifile.Messages {
			msgs = append(msgs, ApiFileMessage{
				File:    &file,
				Message: msg,
			})
		}
	}
	return msgs
}

func showVPPAPIMessages(out io.Writer, msgs []ApiFileMessage, withFields bool) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"File", "Message", "Fields", "Options",
	})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetColumnSeparator(" ")
	table.SetHeaderLine(false)
	table.SetBorder(false)

	for _, msg := range msgs {
		fileName := ""
		if msg.File != nil {
			fileName = msg.File.Name
		}
		msgFields := fmt.Sprintf("%d", len(msg.Fields))
		if withFields {
			msgFields = strings.TrimSpace(yamlTmpl(msgFieldsStr(msg.Fields)))
		}
		msgOptions := getMsgOptionsShort(msg.Options)

		row := []string{
			fileName,
			msg.Name,
			msgFields,
			msgOptions,
		}
		table.Append(row)
	}
	table.Render()
}

func showVPPAPIRPCs(out io.Writer, apifiles []vppapi.File) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"API", "Request", "Reply", "Stream", "StreamMsg", "Events",
	})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	for _, apifile := range apifiles {
		if apifile.Service == nil {
			continue
		}
		for _, rpc := range apifile.Service.RPCs {

			rpcEvents := strings.Join(rpc.Events, ", ")

			var row []string
			row = []string{
				apifile.Name, rpc.Request, rpc.Reply, fmt.Sprint(rpc.Stream), rpc.StreamMsg, rpcEvents,
			}
			table.Append(row)
		}
	}
	table.Render()
}

func msgFieldsStr(fields []vppapi.Field) []string {
	var list []string
	for _, f := range fields {
		var s string
		meta := ""
		if len(f.Meta) != 0 {
			var metas []string
			for k, v := range f.Meta {
				metas = append(metas, fmt.Sprintf("%s=%v", k, v))
			}
			meta = fmt.Sprintf("(%v)", strings.Join(metas, ", "))
		}
		if f.Array {
			length := fmt.Sprint(f.Length)
			if f.SizeFrom != "" {
				length = f.SizeFrom
			}
			s = fmt.Sprintf("%s [%v]%s", f.Name, length, f.Type)
		} else {
			s = fmt.Sprintf("%s %s", f.Name, f.Type)
		}
		if meta != "" {
			s += fmt.Sprintf(" %v", meta)
		}
		list = append(list, s)
	}
	return list
}

func listImportedFiles(apifiles []vppapi.File, apifile *vppapi.File) []string {
	var list []string
	for _, f := range binapigen.ListImportedFiles(apifiles, apifile) {
		list = append(list, f.Name)
	}
	return list
}
