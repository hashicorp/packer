package packer_test

func (ts *PackerTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install plugin with metadata in version", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", BuildSimplePlugin("1.0.0+metadata", ts.T()), "github.com/hashicorp/tester").
			Assert(MustSucceed(), Grep("Successfully installed plugin", grepStdout))
	})

	ts.Run("metadata plugin installed must not have metadata in its path", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+]", grepStdout))
	})

	ts.Run("plugin with metadata should work with validate", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("validate", "./templates/simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})

	ts.Run("plugin with metadata should work with build", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("build", "./templates/simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})
}

func (ts *PackerTestSuite) TestInstallPluginWithPath() {
	tempPluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install plugin with pre-release only", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", BuildSimplePlugin("1.0.0-dev", ts.T()), "github.com/hashicorp/tester").
			Assert(MustSucceed(), Grep("Successfully installed plugin", grepStdout))
	})

	ts.Run("install same plugin with pre-release + metadata", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", BuildSimplePlugin("1.0.0-dev+metadata", ts.T()), "github.com/hashicorp/tester").
			Assert(MustSucceed(), Grep("Successfully installed plugin", grepStdout))
	})

	ts.Run("list plugins, should only report one plugin", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(),
				Grep("plugin-tester_v1.0.0-dev[^+]", grepStdout),
				Grep("plugin-tester_v1.0.0-dev\\+", grepStdout, grepInvert),
				LineCountCheck(1))
	})
}

func (ts *PackerTestSuite) TestInstallPluginPrerelease() {
	pluginPath := BuildSimplePlugin("1.0.1-alpha1", ts.T())

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try install plugin with alpha1 prerelease - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "install", "--path", pluginPath, "github.com/hashicorp/tester").
			Assert(MustFail(), Grep("Packer can only install plugin releases with this command", grepStdout))
	})
}

func (ts *PackerTestSuite) TestRemoteInstallWithPluginsInstall() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install latest version of a remote plugin with packer plugins install", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups").
			Assert(MustSucceed())
	})
}

func (ts *PackerTestSuite) TestRemoteInstallOfPreReleasePlugin() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try to init with a pre-release constraint - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "templates/pre-release_constraint.pkr.hcl").
			Assert(MustFail(),
				Grep("Invalid version constraint", grepStdout),
				Grep("Unsupported prerelease for constraint", grepStdout))
	})

	ts.Run("try to plugins install with a pre-release version - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugin", "install", "github.com/hashicorp/hashicups", "1.0.2-dev").
			Assert(MustFail(),
				Grep("Unsupported prerelease for constraint", grepStdout))
	})
}
