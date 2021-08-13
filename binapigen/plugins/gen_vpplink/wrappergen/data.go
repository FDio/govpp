package wrappergen

import (
	"git.fd.io/govpp.git/binapigen"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

type vppAPIInfo struct {
	APIFile    string
	APIVersion string
	VersionCrc string
}

type Data struct {
	vppAPIInfoMap map[string]*vppAPIInfo
	BinAPI        string
	PackageName   string
}

// NewData creates a new Data struct
//    binAPIPackage - the golang package for the binapi being used
//                    Example: "git.fd.io/govpp.git/binapi"
//    packageName - the short package name for use in generated templates.  Example: vpplink
//    outputDir - the directory to which generated code should be output
func NewDataFromFiles(binAPIPackage, packageName string, files []*binapigen.File) *Data {
	data := &Data{
		BinAPI:        binAPIPackage,
		PackageName:   packageName,
		vppAPIInfoMap: make(map[string]*vppAPIInfo),
	}
	for _, file := range files {
		data.vppAPIInfoMap[file.Desc.Name] = &vppAPIInfo{
			APIFile:    file.Desc.Path,
			APIVersion: file.Version,
			VersionCrc: file.Desc.CRC,
		}
	}
	return data
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
