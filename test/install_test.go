package test

func (ts *PackerTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir, cleanup := ts.MakePluginDir("1.0.0+metadata")
	defer cleanup()

	ts.Run("metadata plugin installed must not have metadata in its path", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(ts.T(), MustSucceed{}, Grep("packer-plugin-tester_v1.0.0[^+]", grepStdout))
	})

	ts.Run("plugin with metadata should work with validate", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("validate", "./templates/simple.pkr.hcl").
			Assert(ts.T(), MustSucceed{}, Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})

	ts.Run("plugin with metadata should work with build", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("build", "./templates/simple.pkr.hcl").
			Assert(ts.T(), MustSucceed{}, Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})
}

func (ts *PackerTestSuite) TestInstallPluginPrerelease() {
	pluginPath := BuildSimplePlugin("1.0.1-alpha1", ts.T())

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try install plugin with alpha1 prerelease - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "install", "--path", pluginPath, "github.com/hashicorp/tester").
			Assert(ts.T(), MustFail{}, Grep("Packer can only install plugin releases with this command", grepStdout))
	})
}
