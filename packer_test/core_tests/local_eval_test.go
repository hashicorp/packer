package core_test

import (
	"fmt"

	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerCoreTestSuite) TestEvalLocalsOrder() {
	ts.SkipNoAcc()

	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		Runs(10).
		Stdin("local.test_local\n").
		SetArgs("console", "./templates/locals_no_order.pkr.hcl").
		Assert(check.MustSucceed(),
			check.GrepInverted("\\[\\]", check.GrepStdout))
}

func (ts *PackerCoreTestSuite) TestLocalDuplicates() {
	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	for _, cmd := range []string{"console", "validate", "build"} {
		ts.Run(fmt.Sprintf("duplicate local detection with %s command - expect error", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "./templates/locals_duplicate.pkr.hcl").
				Assert(check.MustFail(),
					check.Grep("Duplicate local definition"),
					check.Grep("Local variable \"test\" is defined twice"))
		})
	}
}
