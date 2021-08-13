wrappergen is a library to enable generation of wrappers around binapi from Go templates.

It consists of two parts:

1. **wrappergen.Data** - a 'data' interface{} suitable for use with  go [text/template](https://pkg.go.dev/text/template)
   that provides a `````{{ .RequirementSatisfied }}``` method suitable for determining whether the binapi meets version
   requirements.
2. **wrappergen.Template** - An analog of [text/template](https://pkg.go.dev/text/template) for handling a directory
   structure of templates

### wrappergen.Data

wrappergen.Data provides the following for use in templates:

```{{ .RequirementSatisfied }}``` which takes a list of strings for each (APIFile,VersionConstraint) tuple and returns a bool.
Example:
```{{ if .RequirementSatisfied "ipip" ">= 2.0.2" "ip_types" ">= 3.0.0" "interface_types" ">= 1.0.0" }}```

```{{ .PackageName }}``` - Package name for the top of the generated directory
Example:
```package {{ .PackageName }}```

```{{.BinAPI}}``` - Package of the binapi generated against.
Example:
```import "{{ .BinAPI }}/interface_types"```

```{{.PackagePrefix}}``` - Package prefix for the output directory, primarily for use for importing packages generated in subdirectories
Example:
```import "{{ .PackagePrefix }}/types"```

### wrappergen.Template

wrappergen.Template provides a [text/template](https://pkg.go.dev/text/template) analog to ease generating an entire 
directory structure full of files.  It preserves the directory structure in doing so.

Example:

```go
//go:embed templates/*
var templates embed.FS
func main() {
    templates, _ := fs.Sub(templates, "templates")
    tmpl, _ := wrappergen.ParseFS(templates, "*.tmpl")
    data, _ := wrappergen.NewData(binapiPkg, packageName, outputDir)
    tmpl.ExecuteAll(*outputDir, data)
}
```

would process a set of embedded templates for a given binapiPkg to outputDir.

See a full example in [testdata/vpplink/cmd/main.go] and [testdata/consumer/gen.go]
