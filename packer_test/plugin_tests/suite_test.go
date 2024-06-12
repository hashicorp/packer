package plugin_tests

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/lib"
	"github.com/stretchr/testify/suite"
)

type PackerPluginTestSuite struct {
	*lib.PackerTestSuite
}

func Test_PackerPluginSuite(t *testing.T) {
	baseSuite, cleanup := lib.PackerCoreSuite(t)
	defer cleanup()

	ts := &PackerPluginTestSuite{
		baseSuite,
	}
	ts.PackerTestSuite.CompileTestPluginVersions(t,
		"1.0.0",
		"1.0.0",
		"1.0.0+metadata",
		"1.0.1-alpha1",
		"1.0.9",
		"1.0.10",
		"1.0.0-dev",
		"1.0.0-dev+metadata",
		"1.0.10+metadata",
		"1.0.1-dev",
		"2.0.0",
	)

	suite.Run(t, ts)
}
