package core_test

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/common"
	"github.com/stretchr/testify/suite"
)

type PackerCoreTestSuite struct {
	*common.PackerTestSuite
}

func Test_PackerCoreSuite(t *testing.T) {
	baseSuite, cleanup := common.InitBaseSuite(t)
	defer cleanup()

	ts := &PackerCoreTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
