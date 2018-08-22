package api

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestUnspecified(t *testing.T) {
	RegisterTestingT(t)

	var err error = VPPApiError(-1)
	errstr := err.Error()
	Expect(errstr).Should(BeEquivalentTo("VPPApiError: Unspecified Error (-1)"))
}

func TestUnknown(t *testing.T) {
	RegisterTestingT(t)

	var err error = VPPApiError(-999)
	errstr := err.Error()
	Expect(errstr).Should(BeEquivalentTo("VPPApiError: -999"))
}
