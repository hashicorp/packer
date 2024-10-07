package main

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/lib"
	"github.com/stretchr/testify/suite"
)

type PackerDAGTestSuite struct {
	*lib.PackerTestSuite
}

func Test_PackerDAGSuite(t *testing.T) {
	baseSuite, cleanup := lib.PackerCoreSuite(t)
	defer cleanup()

	ts := &PackerDAGTestSuite{
		baseSuite,
	}

	suite.Run(t, ts)
}
