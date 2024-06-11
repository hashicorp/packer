package packer_test

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
