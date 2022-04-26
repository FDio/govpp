package wrappergen_test

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"git.fd.io/govpp.git/wrappergen"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/vpplink/cmd/templates
var templates embed.FS

func TestRequirementSatisfied(t *testing.T) {
	outputDir := "./testdata/output"
	data, err := wrappergen.NewData("git.fd.io/govpp.git/binapi", "", outputDir)
	require.NoError(t, err)
	//data.RequirementSatisfied("ip_types", ">= 3.0.0", "interface_types", ">= 1.0.0", "ipip", ">= 2.0.2")
	require.True(t, data.RequirementSatisfied("ipip", ">= 2.0.2"))
	require.False(t, data.RequirementSatisfied("ipip", "< 2.0.2"))
}

func TestReqirementSatisfiedInTemplate(t *testing.T) {
	outputDir := "./testdata/output"
	data, err := wrappergen.NewData("git.fd.io/govpp.git/binapi", "", outputDir)
	require.NoError(t, err)
	templates, err := fs.Sub(templates, "testdata/cmd/templates")
	require.NoError(t, err)
	tmpl, err := wrappergen.ParseFS(templates, "*.tmpl")
	require.NoError(t, err)
	err = tmpl.ExecuteAll(outputDir, data)
	require.NoError(t, err)
	require.NoError(t, os.RemoveAll(outputDir))
}
