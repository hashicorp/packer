package gob_test

import (
	"testing"

	"github.com/hashicorp/packer/packer_test/lib"
	"github.com/stretchr/testify/suite"
)

type PackerGobTestSuite struct {
	*lib.PackerTestSuite
}

func Test_PackerPluginSuite(t *testing.T) {
	baseSuite, cleanup := lib.PackerCoreSuite(t)
	defer cleanup()

	ts := &PackerGobTestSuite{
		baseSuite,
	}

	// Build two versions of each plugin, one with gob only, one with protobuf only
	//
	// We'll install them manually in tests, as they'll need to be installed as
	// different plugin sources in order for discovery to trigger the
	// gob-only/pb-supported behaviours we want to test.
	ts.BuildSimplePlugin("1.1.0+pb", t, lib.UseDependency(lib.SDKModule, "grpc_base"))
	ts.BuildSimplePlugin("1.0.0+pb", t, lib.UseDependency(lib.SDKModule, "grpc_base"))

	ts.BuildSimplePlugin("1.0.0+gob", t)
	ts.BuildSimplePlugin("1.1.0+gob", t)

	suite.Run(t, ts)
}
