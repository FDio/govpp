package main

import (
	"embed"
	_ "embed"
	"flag"
	"io/fs"
	"os"

	"git.fd.io/govpp.git/wrappergen"
	"github.com/sirupsen/logrus"
)

//go:embed templates/*
var templates embed.FS

func main() {
	// Parse flags
	binapiPkg := flag.String("binapi-package", "git.fd.io/govpp.git/binapi", "BinAPI Package to generate from")
	outputDir := flag.String("output-dir", ".", "Output directory for generated files.")
	flag.Parse()

	// Get the package from the standard env variable used by go:generate
	packageName := os.Getenv("GOPACKAGE")

	data, err := wrappergen.NewData(*binapiPkg, packageName, *outputDir)
	if err != nil {
		logrus.Fatalf("error creating wrappergen.Data for binapiPkg: %s packageName %s", *binapiPkg, packageName)
	}

	// Trim off the "templates" prefix from the paths of our templates
	templates, err := fs.Sub(templates, "templates")
	if err != nil {
		logrus.Fatalf("error creating subFS for 'templates'")
	}

	// Parse all the templates
	tmpl, err := wrappergen.ParseFS(templates, "*.tmpl")
	if err != nil {
		logrus.Fatalf("failed to ParseFS templates: %s", err)
	}

	// Execute all the templates
	if err := tmpl.ExecuteAll(*outputDir, data); err != nil {
		logrus.Fatalf("failed to execute template: %s", err)
	}
}
