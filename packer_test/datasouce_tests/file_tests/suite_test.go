package main

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/common"
	"github.com/stretchr/testify/suite"
)

type FileDatasourceTestSuite struct {
	*common.PackerTestSuite
}

func Test_FileDatasourceTestSuite(t *testing.T) {
	baseSuite, cleanup := common.InitBaseSuite(t)
	defer cleanup()

	ts := &FileDatasourceTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
