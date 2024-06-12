package core_test

import (
	"fmt"

	"github.com/hashicorp/packer/packer_test/lib"
)

func (ts *PackerCoreTestSuite) TestEvalLocalsOrder() {
	ts.SkipNoAcc()

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		Runs(10).
		Stdin("local.test_local\n").
		SetArgs("console", "./templates/locals_no_order.pkr.hcl").
		Assert(lib.MustSucceed(),
			lib.Grep("\\[\\]", lib.GrepStdout, lib.GrepInvert))
}

func (ts *PackerCoreTestSuite) TestLocalDuplicates() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	for _, cmd := range []string{"console", "validate", "build"} {
		ts.Run(fmt.Sprintf("duplicate local detection with %s command - expect error", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "./templates/locals_duplicate.pkr.hcl").
				Assert(lib.MustFail(),
					lib.Grep("Duplicate local definition"),
					lib.Grep("Local variable \"test\" is defined twice"))
		})
	}
}
