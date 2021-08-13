package wrappergen

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Template struct {
	templates map[string]*template.Template
	input     fs.FS
}

func ParseFS(input fs.FS, patterns ...string) (*Template, error) {
	rv := &Template{
		input:     input,
		templates: make(map[string]*template.Template),
	}
	var err error
	fs.WalkDir(rv.input, ".", rv.addAllToTemplateWalkFn)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (t *Template) addAllToTemplateWalkFn(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}
	if !strings.HasSuffix(d.Name(), ".tmpl") {
		return nil
	}
	tmpl, err := template.ParseFS(t.input, path)
	if err != nil {
		return err
	}
	t.templates[path] = tmpl
	return nil
}

func (t *Template) createExecuteWalkFn(outputDir string, data interface{}) func(path string, d fs.DirEntry, err error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".tmpl") {
			return nil
		}

		outputBuffer := bytes.NewBuffer([]byte{})
		tmpl, ok := t.templates[path]
		if !ok {
			return nil
		}
		if err := tmpl.Execute(outputBuffer, data); err != nil {
			return err
		}

		if strings.TrimSpace(outputBuffer.String()) == "" {
			return nil
		}

		outputPath := filepath.Join(outputDir, path)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
			return err
		}

		output, err := os.Create(strings.TrimSuffix(outputPath, ".tmpl"))
		defer output.Close()

		_, err = io.Copy(output, outputBuffer)
		return err
	}
}

func (t *Template) ExecuteAll(outputDir string, data interface{}) error {
	return fs.WalkDir(t.input, ".", t.createExecuteWalkFn(outputDir, data))
}
