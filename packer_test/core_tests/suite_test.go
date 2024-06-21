package core_test

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/lib"
	"github.com/stretchr/testify/suite"
)

type PackerCoreTestSuite struct {
	*lib.PackerTestSuite
}

func Test_PackerPluginSuite(t *testing.T) {
	baseSuite, cleanup := lib.PackerCoreSuite(t)
	defer cleanup()

	ts := &PackerCoreTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
