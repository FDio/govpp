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
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
)

type VppApiCmdOptions struct {
	Input           string
	Format          string
	IncludeImported bool
	IncludeFields   bool

	ShowContents bool
	ShowMessages bool
	ShowRPC      bool
	ShowRaw      bool
}

func newVppapiCmd() *cobra.Command {
	var (
		opts = VppApiCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:   "vppapi",
		Short: "Browse VPP API",
		Long:  "Browse VPP API files",
		RunE: func(cmd *cobra.Command, args []string) error {
			listOpts := VppApiListCmdOptions{VppApiCmdOptions: opts}
			return runVppApiListCmd(cmd.OutOrStdout(), listOpts, args)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", opts.Input, "Path to directory containing VPP API files")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "", "Output format (json, yaml, go-template..)")
	cmd.PersistentFlags().BoolVar(&opts.IncludeImported, "include-imported", false, "Include imported types")
	cmd.PersistentFlags().BoolVar(&opts.IncludeFields, "include-fields", false, "Include message fields")

	cmd.AddCommand(
		newVppapiListCmd(),
	)

	return cmd
}

type VppApiListCmdOptions struct {
	VppApiCmdOptions

	ShowContents bool
	ShowMessages bool
	ShowRPC      bool
	ShowRaw      bool
}

func newVppapiListCmd() *cobra.Command {
	var (
		opts = VppApiListCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:     "ls [FILE, ...]",
		Aliases: []string{"l", "list"},
		Short:   "List VPP API files",
		Long:    "List VPP API files and their contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVppApiListCmd(cmd.OutOrStdout(), opts, args)
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.ShowContents, "show-contents", false, "Show contents of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowRPC, "show-rpc", false, "Show service RPCs of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowMessages, "show-messages", false, "Show messages for VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowRaw, "show-raw", false, "Show raw VPP API file(s)")

	return cmd
}

func runVppApiListCmd(out io.Writer, opts VppApiListCmdOptions, args []string) error {
	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Debugf("parsing VPP API directory: %v", vppInput.ApiDirectory)

	// parse API directory
	allapifiles, err := vppapi.ParseDir(vppInput.ApiDirectory)
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
		for i, apifile := range apifiles {
			f := apifile
			binapigen.RemoveImportedTypes(allapifiles, &f)
			apifiles[i] = f
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
		} else if opts.ShowRaw {
			showVPPAPIRaw(out, vppInput, args)
		} else {
			showVPPAPIList(out, apifiles)
		}
	} else {
		if err := formatAsTemplate(out, format, apifiles); err != nil {
			return err
		}
	}

	return nil
}

type VppApiRawFile struct {
	Name    string
	Path    string
	Content string
}

func showVPPAPIRaw(out io.Writer, input *vppapi.VppInput, args []string) {
	list, err := vppapi.FindFiles(input.ApiDirectory)
	if err != nil {
		logrus.Errorf("failed to find files: %v", err)
		return
	}

	var allapifiles []VppApiRawFile
	for _, f := range list {
		file, err := vppapi.ParseFile(f)
		if err != nil {
			logrus.Debugf("failed to parse file: %v", err)
			continue
		}
		// use file path relative to apiDir
		if p, err := filepath.Rel(input.ApiDirectory, file.Path); err == nil {
			file.Path = p
		}
		// extract file name
		base := filepath.Base(f)
		name := base[:strings.Index(base, ".")]
		b, err := os.ReadFile(f)
		if err != nil {
			logrus.Errorf("failed to read file %q: %v", f, err)
			continue
		}

		allapifiles = append(allapifiles, VppApiRawFile{
			Path:    file.Path,
			Name:    name,
			Content: string(b),
		})
	}

	var apifiles []VppApiRawFile
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
				logrus.Errorf("VPP API file %q not found", arg)
				continue

			}
		}
	} else {
		// select all files
		apifiles = allapifiles
	}

	for _, f := range apifiles {
		fmt.Printf("# %s (%v)\n", f.Name, f.Path)
		fmt.Printf("%s\n", f.Content)
		fmt.Println()
	}
}

type VppApiFile struct {
	File        string
	Version     string
	CRC         string
	Options     string
	Path        string
	NumImports  uint
	NumMessages uint
	NumTypes    uint
	NumRPCs     uint
}

func showVPPAPIList(out io.Writer, apifiles []vppapi.File) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "API", "Version", "CRC", "Path", "Imports", "Messages", "Types", "RPCs", "Options",
	})
	//table.SetAutoMergeCells(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_RIGHT, 0, 0, tablewriter.ALIGN_LEFT, 0, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT,
	})

	for i, apifile := range apifiles {
		index := i + 1

		typesCount := getFileTypesCount(apifile)
		apiVersion := getFileVersion(apifile)
		apiCrc := getShortCrc(apifile.CRC)
		options := strings.Join(getFileOptionsSlice(apifile), ", ")
		apiDirPath := path.Dir(pathOfParentAndFile(apifile.Path))

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
		var importCount string
		if len(apifile.Imports) > 0 {
			importCount = fmt.Sprintf("%2d", len(apifile.Imports))
		} else {
			importCount = fmt.Sprintf("%2s", "-")
		}

		row := []string{
			fmt.Sprint(index), apifile.Name, apiVersion, apiCrc, apiDirPath, importCount, msgs, types, services, options,
		}

		table.Append(row)
	}
	table.Render()
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
		apiVersion := getFileVersion(apifile)
		options := getFileOptionsSlice(apifile)

		// print file header
		fmt.Fprintln(&buf, "--------------------------------------------------------------------------------")
		fmt.Fprintf(&buf, "# %v (%s) %v (%v) \n", apifile.Name, apifile.Path, apiVersion, apifile.CRC)
		if len(options) > 0 {
			fmt.Fprintf(&buf, "# Options: %v\n", mapStr(apifile.Options))
		}
		fmt.Fprintln(&buf, "--------------------------------------------------------------------------------")
		fmt.Fprintln(&buf)

		// Messages
		if len(apifile.Messages) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "Message", "CRC", "Fields", "Options", "Comment"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoWrapText(false)
			table.SetRowLine(false)
			table.SetColumnSeparator(" ")
			table.SetHeaderLine(false)
			table.SetBorder(false)

			for i, msg := range apifile.Messages {
				index := i + 1
				msgCrc := getShortCrc(msg.CRC)
				msgFields := fmt.Sprintf("%v", len(msg.Fields))
				msgOptions := shorMessageOptions(msg.Options)
				msgComment := normalizeMessageComment(msg.Comment)
				row := []string{strconv.Itoa(index), msg.Name, msgCrc, msgFields, msgOptions, msgComment}
				table.Append(row)
			}

			table.Render()
			buf.Write([]byte("\n"))
		}

		// Structs
		if len(apifile.StructTypes) > 0 {
			table := tablewriter.NewWriter(&buf)
			table.SetHeader([]string{"#", "Type", "Fields"})
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
			table.SetHeader([]string{"#", "Enum", "Type", "Kind", "Entries"})
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
			table.SetHeader([]string{"#", "Alias", "Type", "Length"})
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

func getShortCrc(crc string) string {
	return strings.TrimPrefix(crc, "0x")
}

const maxLengthMsgOptionValue = 30

func shorMessageOptions(options map[string]string) string {
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

func normalizeMessageComment(comment string) string {
	if comment == "" {
		return ""
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

type VppApiFileMessage struct {
	File *vppapi.File
	vppapi.Message
}

func listVPPAPIMessages(apifiles []vppapi.File) []VppApiFileMessage {
	var msgs []VppApiFileMessage
	for _, apifile := range apifiles {
		file := apifile
		for _, msg := range apifile.Messages {
			msgs = append(msgs, VppApiFileMessage{
				File:    &file,
				Message: msg,
			})
		}
	}
	return msgs
}

func showVPPAPIMessages(out io.Writer, msgs []VppApiFileMessage, withFields bool) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "File", "Message", "Fields", "Options",
	})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetColumnSeparator(" ")
	table.SetHeaderLine(false)
	table.SetBorder(false)

	for i, msg := range msgs {
		index := fmt.Sprint(i + 1)
		fileName := ""
		if msg.File != nil {
			fileName = msg.File.Name
		}
		msgFields := fmt.Sprintf("%d", len(msg.Fields))
		if withFields {
			msgFields = strings.TrimSpace(yamlTmpl(msgFieldsStr(msg.Fields)))
		}
		msgOptions := shorMessageOptions(msg.Options)

		row := []string{
			index, fileName, msg.Name, msgFields, msgOptions,
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

			row := []string{
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

func getFileTypesCount(apifile vppapi.File) int {
	return len(apifile.EnumTypes) + len(apifile.EnumflagTypes) + len(apifile.AliasTypes) + len(apifile.StructTypes) + len(apifile.UnionTypes)
}

func getFileVersion(apifile vppapi.File) string {
	for k, v := range apifile.Options {
		if k == vppapi.OptFileVersion {
			return v
		}
	}
	return "0.0.0"
}

func getFileOptionsSlice(apifile vppapi.File) []string {
	var options []string
	for k, v := range apifile.Options {
		if k == vppapi.OptFileVersion {
			continue
		}
		options = append(options, fmt.Sprintf("%s=%v", k, v))
	}
	return options
}
