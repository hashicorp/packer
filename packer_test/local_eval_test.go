package packer_test

import "fmt"

func (ts *PackerTestSuite) TestEvalLocalsOrder() {
	ts.SkipNoAcc()

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		Runs(10).
		Stdin("local.test_local\n").
		SetArgs("console", "./templates/locals_no_order.pkr.hcl").
		Assert(MustSucceed(), Grep("\\[\\]", grepStdout, grepInvert))
}

func (ts *PackerTestSuite) TestLocalDuplicates() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	for _, cmd := range []string{"console", "validate", "build"} {
		ts.Run(fmt.Sprintf("duplicate local detection with %s command - expect error", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "./templates/locals_duplicate.pkr.hcl").
				Assert(MustFail(),
					Grep("Duplicate local definition"),
					Grep("Local variable \"test\" is defined twice"))
		})
	}
}
