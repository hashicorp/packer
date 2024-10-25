package plugin_tests

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/common"
	"github.com/stretchr/testify/suite"
)

type PackerHCPSbomTestSuite struct {
	*common.PackerTestSuite
}

func Test_PackerPluginSuite(t *testing.T) {
	baseSuite, cleanup := common.InitBaseSuite(t)
	defer cleanup()

	ts := &PackerHCPSbomTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
