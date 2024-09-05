package main

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/common"
	"github.com/stretchr/testify/suite"
)

type PackerDAGTestSuite struct {
	*common.PackerTestSuite
}

func Test_PackerDAGSuite(t *testing.T) {
	baseSuite, cleanup := common.InitBaseSuite(t)
	defer cleanup()

	ts := &PackerDAGTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
