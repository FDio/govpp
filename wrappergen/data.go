package wrappergen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

type vppAPIInfo struct {
	APIFile    string
	APIVersion string
	VersionCrc string
}

type Data struct {
	binAPIFS      fs.FS
	vppAPIInfoMap map[string]*vppAPIInfo
	binAPIPath    string
	BinAPI        string
	PackageName   string
	PackagePrefix string
}

// NewData creates a new Data struct
//    binAPIPackage - the golang package for the binapi being used
//                    Example: "git.fd.io/govpp.git/binapi"
//    packageName - the short package name for use in generated templates.  Example: vpplink
//    outputDir - the directory to which generated code should be output
func NewData(binAPIPackage, packageName, outputDir string) (*Data, error) {
	rv := &Data{
		vppAPIInfoMap: make(map[string]*vppAPIInfo),
		PackageName:   packageName,
		BinAPI:        binAPIPackage,
	}

	// Get directory containing the binAPI package
	pathBytes, err := exec.Command("go", "list", "-f", "{{ .Dir }}", binAPIPackage).Output()
	if err != nil {
		return nil, err
	}
	rv.binAPIPath = string(pathBytes)
	rv.binAPIPath = strings.TrimSpace(rv.binAPIPath)
	rv.binAPIFS = os.DirFS(rv.binAPIPath)

	// Extract the PackagePrefix from outputDir
	rv.PackagePrefix, err = resolveImportPath(outputDir)
	if err != nil {
		return nil, err
	}

	err = fs.WalkDir(rv.binAPIFS, ".", func(path string, d fs.DirEntry, err error) error {
		fileSet := token.NewFileSet()
		if err != nil {
			return err
		}
		if d == nil || d.IsDir() {
			return nil
		}

		file, err := rv.binAPIFS.Open(path)
		if err != nil {
			return err
		}
		astFile, err := parser.ParseFile(fileSet, path, file, 0)
		if err != nil {
			return err
		}

		if astFile == nil || astFile.Scope == nil {
			return nil
		}
		info := &vppAPIInfo{
			APIFile:    extractConstantValue(astFile, "APIFile"),
			APIVersion: extractConstantValue(astFile, "APIVersion"),
			VersionCrc: extractConstantValue(astFile, "VersionCrc"),
		}
		if info.APIFile == "" || info.APIVersion == "" || info.VersionCrc == "" {
			return nil
		}
		rv.vppAPIInfoMap[astFile.Name.Name] = info

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (d *Data) RequirementSatisfied(reqs ...string) bool {
	if len(reqs)%2 != 0 {
		return false
	}

	for i := 0; i < len(reqs)/2; i++ {
		apiName := reqs[2*i]
		versionConstraint := reqs[2*i+1]
		info, ok := d.vppAPIInfoMap[apiName]
		if !ok {
			return false
		}
		v, err := version.NewVersion(info.APIVersion)
		if err != nil {
			logrus.Debugf("Could not parse %s api version: %s", apiName, versionConstraint)
		}
		constraint, err := version.NewConstraint(versionConstraint)
		if err != nil {
			logrus.Debugf("Could not parse template %s dependency version constraint: %s", apiName, versionConstraint)
		}
		if !constraint.Check(v) {
			return false
		}
	}

	return true
}

func extractConstantValue(file *ast.File, name string) string {
	if file == nil || file.Scope == nil {
		return ""
	}
	con := file.Scope.Lookup(name)
	if con == nil || con.Kind != ast.Con {
		return ""
	}
	valueSpec, ok := con.Decl.(*ast.ValueSpec)
	if !ok {
		return ""
	}
	if len(valueSpec.Values) != 1 {
		return ""
	}
	value, ok := valueSpec.Values[0].(*ast.BasicLit)
	if !ok {
		return ""
	}
	if value.Kind == token.STRING {
		unquotedValue, err := strconv.Unquote(value.Value)
		if err != nil {
			return ""
		}
		return unquotedValue
	}
	return value.Value
}

// resolveImportPath tries to resolve import path for a directory.
func resolveImportPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	modRoot := findGoModuleRoot(absPath)
	if modRoot == "" {
		return "", err
	}
	modPath, err := readModulePath(path.Join(modRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	relDir, err := filepath.Rel(modRoot, absPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(modPath, relDir), nil
}

// findGoModuleRoot looks for enclosing Go module.
func findGoModuleRoot(dir string) (root string) {
	dir = filepath.Clean(dir)
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

var modulePathRE = regexp.MustCompile(`module[ \t]+([^ \t\r\n]+)`)

// readModulePath reads module path from go.mod file.
func readModulePath(gomod string) (string, error) {
	data, err := ioutil.ReadFile(gomod)
	if err != nil {
		return "", err
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return "", err
	}
	return string(m[1]), nil
}
