package vscodiumbuildpack_test

import (
	"testing"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitVSCodium(t *testing.T) {
	format.MaxLength = 0
	suite := spec.New("vscodium", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite.Run(t)
}
