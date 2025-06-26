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

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

const exampleVppApiLsCmd = `
  <cyan># List all VPP API files</>
  govpp vppapi ls [INPUT]

  <cyan># List only files matching path(s)</>
  govpp vppapi ls [INPUT] --path "core/"
  govpp vppapi ls [INPUT] --path "vpe,memclnt"

  <cyan># List all contents from files</>
  govpp vppapi ls [INPUT] --show-contents

  <cyan># List message definitions</>
  govpp vppapi ls [INPUT] --show-messages
  govpp vppapi ls [INPUT] --show-messages --include-fields

  <cyan># Print raw VPP API files</>
  govpp vppapi ls [INPUT] --show-raw
`

type VppApiLsCmdOptions struct {
	*VppApiCmdOptions

	Format string

	IncludeImported bool
	IncludeFields   bool

	ShowContents bool
	ShowMessages bool
	ShowRPC      bool
	ShowRaw      bool

	SortByName bool
}

func newVppApiLsCmd(cli Cli, vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiLsCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:     "list [INPUT] [--path PATH]... [--show-contents | --show-messages | --show-raw | --show-rpc]",
		Aliases: []string{"l", "ls"},
		Short:   "List VPP API contents",
		Long:    "List VPP API files and their contents",
		Example: color.Sprint(exampleVppApiLsCmd),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			return runVppApiLsCmd(cli.Out(), opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "", "Format for the output (json, yaml, go-template..)")
	cmd.PersistentFlags().BoolVar(&opts.IncludeImported, "include-imported", false, "Include imported types")
	cmd.PersistentFlags().BoolVar(&opts.IncludeFields, "include-fields", false, "Include message fields")
	cmd.PersistentFlags().BoolVar(&opts.ShowContents, "show-contents", false, "Show contents of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowRPC, "show-rpc", false, "Show service RPCs of VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowMessages, "show-messages", false, "Show messages for VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.ShowRaw, "show-raw", false, "Show raw VPP API file(s)")
	cmd.PersistentFlags().BoolVar(&opts.SortByName, "sort-by-name", true, "Sort by name")

	return cmd
}

func runVppApiLsCmd(out io.Writer, opts VppApiLsCmdOptions) error {
	vppInput, err := resolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	apifiles, err := prepareVppApiFiles(vppInput.Schema.Files, opts.Paths, opts.IncludeImported, opts.SortByName)
	if err != nil {
		return err
	}

	// format output
	if opts.Format == "" {
		if opts.ShowMessages {
			apimsgs := listVPPAPIMessages(apifiles)
			return showVPPAPIMessages(out, apimsgs, opts.IncludeFields)
		} else if opts.ShowRPC {
			return showVPPAPIRPCs(out, apifiles)
		} else if opts.ShowContents {
			return showVPPAPIContents(out, apifiles)
		} else if opts.ShowRaw {
			rawFiles := getVppApiRawFiles(vppInput.ApiDirectory, apifiles)
			return showVPPAPIRaw(out, rawFiles)
		} else {
			files := vppApiFilesToList(apifiles)
			return showVPPAPIFilesTable(out, files)
		}
	} else {
		if err := formatAsTemplate(out, opts.Format, apifiles); err != nil {
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

func getVppApiRawFiles(apiDir string, files []vppapi.File) []VppApiRawFile {
	var rawFiles []VppApiRawFile
	for _, file := range files {
		filePath := filepath.Join(apiDir, file.Path)
		b, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Errorf("failed to read file %q: %v", filePath, err)
			continue
		}
		rawFiles = append(rawFiles, VppApiRawFile{
			Path:    file.Path,
			Name:    file.Name,
			Content: string(b),
		})
	}
	return rawFiles
}

func showVPPAPIRaw(out io.Writer, rawFiles []VppApiRawFile) error {
	if len(rawFiles) == 0 {
		logrus.Errorf("no files to show")
		return nil
	}
	for _, f := range rawFiles {
		fmt.Fprintf(out, "# %s (%v)\n", f.Name, f.Path)
		fmt.Fprintf(out, "%s\n", f.Content)
		fmt.Fprintln(out)
	}
	return nil
}

// VppApiFile holds info about a single VPP API file used for listing files.
type VppApiFile struct {
	Name        string
	Version     string
	CRC         string
	Options     []string
	Path        string
	NumImports  int
	NumMessages int
	NumTypes    int
	NumRPCs     int
}

func showVPPAPIFilesTable(out io.Writer, apifiles []VppApiFile) error {
	cfg := tablewriter.NewConfigBuilder()
	cfg.Row().Alignment().WithPerColumn([]tw.Align{
		tw.AlignRight, tw.AlignNone, tw.AlignNone, tw.AlignLeft, tw.AlignNone,
		tw.AlignRight, tw.AlignRight, tw.AlignRight, tw.AlignRight, tw.AlignLeft,
	})
	cfg.WithRowAutoWrap(tw.WrapNone)
	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Off},
			},
		}),
		tablewriter.WithConfig(cfg.Build()),
	)

	table.Header(
		"#", "API", "Version", "CRC", "Path",
		"Imports", "Messages", "Types", "RPCs", "Options",
	)
	for i, apifile := range apifiles {
		index := i + 1

		typesCount := apifile.NumTypes
		apiVersion := apifile.Version
		apiCrc := apifile.CRC
		options := strings.Join(apifile.Options, ", ")
		apiDirPath := path.Dir(apifile.Path)

		var msgs string
		if apifile.NumMessages > 0 {
			msgs = fmt.Sprintf("%3d", apifile.NumMessages)
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
		if apifile.NumRPCs > 0 {
			services = fmt.Sprintf("%2d", apifile.NumRPCs)
		} else {
			services = fmt.Sprintf("%2s", "-")
		}
		var importCount string
		if apifile.NumImports > 0 {
			importCount = fmt.Sprintf("%2d", apifile.NumImports)
		} else {
			importCount = fmt.Sprintf("%2s", "-")
		}

		err := table.Append(
			fmt.Sprint(index), apifile.Name, apiVersion, apiCrc, apiDirPath,
			importCount, msgs, types, services, options,
		)
		if err != nil {
			return err
		}
	}
	return table.Render()
}

func vppApiFilesToList(apifiles []vppapi.File) []VppApiFile {
	var list []VppApiFile
	for _, apifile := range apifiles {
		numRPCs := 0
		if apifile.Service != nil {
			numRPCs = len(apifile.Service.RPCs)
		}
		list = append(list, VppApiFile{
			Name:        apifile.Name,
			Version:     getFileVersion(apifile),
			CRC:         getShortCrc(apifile.CRC),
			Options:     getFileOptionsSlice(apifile),
			Path:        pathOfParentAndFile(apifile.Path),
			NumImports:  len(apifile.Imports),
			NumMessages: len(apifile.Messages),
			NumTypes:    getFileTypesCount(apifile),
			NumRPCs:     numRPCs,
		})
	}
	return list
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

func showVPPAPIContents(out io.Writer, apifiles []vppapi.File) error {
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
			cfg := tablewriter.NewConfigBuilder()
			cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
			table := tablewriter.NewTable(
				&buf,
				tablewriter.WithRendition(tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Separators: tw.Separators{
							BetweenRows:    tw.Off,
							BetweenColumns: tw.Off,
						},
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
					},
				}),
				tablewriter.WithRowAutoWrap(tw.WrapNone),
				tablewriter.WithConfig(cfg.Build()),
			)

			table.Header("#", "Message", "CRC", "Fields", "Options", "Comment")

			for i, msg := range apifile.Messages {
				index := i + 1
				msgCrc := getShortCrc(msg.CRC)
				msgFields := fmt.Sprintf("%v", len(msg.Fields))
				msgOptions := shorMessageOptions(msg.Options)
				msgComment := normalizeMessageComment(msg.Comment)
				err := table.Append(
					strconv.Itoa(index), msg.Name, msgCrc, msgFields, msgOptions, msgComment,
				)
				if err != nil {
					return err
				}
			}

			err := table.Render()
			if err != nil {
				return err
			}
			buf.Write([]byte("\n"))
		}

		// Structs
		if len(apifile.StructTypes) > 0 {
			cfg := tablewriter.NewConfigBuilder()
			cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
			table := tablewriter.NewTable(
				&buf,
				tablewriter.WithRendition(tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Separators: tw.Separators{
							BetweenRows:    tw.Off,
							BetweenColumns: tw.Off,
						},
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
					},
				}),
				tablewriter.WithRowAutoWrap(tw.WrapNone),
				tablewriter.WithConfig(cfg.Build()),
			)
			table.Header("#", "Type", "Fields")

			for i, typ := range apifile.StructTypes {
				fields := fmt.Sprintf("%v", len(typ.Fields))
				err := table.Append(strconv.Itoa(i+1), typ.Name, fields)
				if err != nil {
					return err
				}
			}

			err := table.Render()
			if err != nil {
				return err
			}
			buf.Write([]byte("\n"))
		}

		// Unions
		if len(apifile.UnionTypes) > 0 {
			cfg := tablewriter.NewConfigBuilder()
			cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
			table := tablewriter.NewTable(
				&buf,
				tablewriter.WithRendition(tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Separators: tw.Separators{
							BetweenRows:    tw.Off,
							BetweenColumns: tw.Off,
						},
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
					},
				}),
				tablewriter.WithRowAutoWrap(tw.WrapNone),
				tablewriter.WithConfig(cfg.Build()),
			)
			table.Header("#", "UNION", "FIELDS")

			for i, typ := range apifile.UnionTypes {
				fields := fmt.Sprintf("%v", len(typ.Fields))
				err := table.Append(strconv.Itoa(i+1), typ.Name, fields)
				if err != nil {
					return err
				}
			}
			err := table.Render()
			if err != nil {
				return err
			}
			buf.Write([]byte("\n"))
		}

		// Enums
		if len(apifile.EnumTypes) > 0 || len(apifile.EnumflagTypes) > 0 {
			cfg := tablewriter.NewConfigBuilder()
			cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
			table := tablewriter.NewTable(
				&buf,
				tablewriter.WithRendition(tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Separators: tw.Separators{
							BetweenRows:    tw.Off,
							BetweenColumns: tw.Off,
						},
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
					},
				}),
				tablewriter.WithRowAutoWrap(tw.WrapNone),
				tablewriter.WithConfig(cfg.Build()),
			)
			table.Header("#", "Enum", "Type", "Kind", "Entries")

			for i, typ := range apifile.EnumTypes {
				typEntries := fmt.Sprintf("%v", len(typ.Entries))
				err := table.Append(
					strconv.Itoa(i+1), typ.Name, typ.Type, "enum", typEntries,
				)
				if err != nil {
					return err
				}
			}
			for i, typ := range apifile.EnumflagTypes {
				typEntries := fmt.Sprintf("%v", len(typ.Entries))
				err := table.Append(
					strconv.Itoa(i+1), typ.Name, typ.Type, "flag", typEntries,
				)
				if err != nil {
					return err
				}
			}

			err := table.Render()
			if err != nil {
				return err
			}
			buf.Write([]byte("\n"))
		}

		// Aliases
		if len(apifile.AliasTypes) > 0 {
			cfg := tablewriter.NewConfigBuilder()
			cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
			table := tablewriter.NewTable(
				&buf,
				tablewriter.WithRendition(tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Separators: tw.Separators{
							BetweenRows:    tw.Off,
							BetweenColumns: tw.Off,
						},
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
					},
				}),
				tablewriter.WithRowAutoWrap(tw.WrapNone),
				tablewriter.WithConfig(cfg.Build()),
			)
			table.Header("#", "Alias", "Type", "Length")

			for i, typ := range apifile.AliasTypes {
				err := table.Append(
					strconv.Itoa(i+1), typ.Name, typ.Type, strconv.Itoa(typ.Length),
				)
				if err != nil {
					return err
				}
			}

			err := table.Render()
			if err != nil {
				return err
			}
			buf.Write([]byte("\n"))

		}

	}

	fmt.Fprint(out, buf.String())
	return nil
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

func showVPPAPIMessages(out io.Writer, msgs []VppApiFileMessage, withFields bool) error {
	cfg := tablewriter.NewConfigBuilder()
	cfg.Header().Alignment().WithGlobal(tw.AlignLeft)
	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows:    tw.Off,
					BetweenColumns: tw.Off,
				},
				Lines: tw.Lines{
					ShowHeaderLine: tw.Off,
				},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithConfig(cfg.Build()),
	)
	table.Header("#", "File", "Message", "Fields", "Options")

	for i, msg := range msgs {
		fileName := ""
		if msg.File != nil {
			fileName = msg.File.Name
		}
		msgFields := fmt.Sprintf("%d", len(msg.Fields))
		if withFields {
			msgFields = strings.TrimSpace(yamlTmpl(msgFieldsStr(msg.Fields)))
		}
		msgOptions := shorMessageOptions(msg.Options)
		err := table.Append(
			fmt.Sprint(i+1), fileName, msg.Name, msgFields, msgOptions,
		)
		if err != nil {
			return err
		}
	}
	return table.Render()
}

func showVPPAPIRPCs(out io.Writer, apifiles []vppapi.File) error {
	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.On},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
	)
	table.Header("API", "Request", "Reply", "Stream", "StreamMsg", "Events")

	for _, apifile := range apifiles {
		if apifile.Service == nil {
			continue
		}
		for _, rpc := range apifile.Service.RPCs {
			rpcEvents := strings.Join(rpc.Events, ", ")
			err := table.Append(
				apifile.Name, rpc.Request, rpc.Reply,
				fmt.Sprint(rpc.Stream), rpc.StreamMsg, rpcEvents,
			)
			if err != nil {
				return err
			}
		}
	}
	return table.Render()
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
