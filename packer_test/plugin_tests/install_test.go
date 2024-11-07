package plugin_tests

import "github.com/hashicorp/packer/packer_test/common/check"

func (ts *PackerPluginTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir := ts.MakePluginDir()
	defer tempPluginDir.Cleanup()

	ts.Run("install plugin with metadata in version", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0+metadata"), "github.com/hashicorp/tester").
			Assert(check.MustSucceed(), check.Grep("Successfully installed plugin", check.GrepStdout))
	})

	ts.Run("metadata plugin installed must not have metadata in its path", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1.0.0[^+]", check.GrepStdout))
	})

	ts.Run("plugin with metadata should work with validate", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("validate", "./templates/simple.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", check.GrepStderr))
	})

	ts.Run("plugin with metadata should work with build", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("build", "./templates/simple.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", check.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestInstallPluginWithPath() {
	tempPluginDir := ts.MakePluginDir()
	defer tempPluginDir.Cleanup()

	ts.Run("install plugin with pre-release only", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0-dev"), "github.com/hashicorp/tester").
			Assert(check.MustSucceed(), check.Grep("Successfully installed plugin", check.GrepStdout))
	})

	ts.Run("install same plugin with pre-release + metadata", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0-dev+metadata"), "github.com/hashicorp/tester").
			Assert(check.MustSucceed(), check.Grep("Successfully installed plugin", check.GrepStdout))
	})

	ts.Run("list plugins, should only report one plugin", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(),
				check.Grep("plugin-tester_v1.0.0-dev[^+]", check.GrepStdout),
				check.GrepInverted("plugin-tester_v1.0.0-dev\\+", check.GrepStdout),
				check.LineCountCheck(1))
	})
}

func (ts *PackerPluginTestSuite) TestInstallPluginPrerelease() {
	pluginPath := ts.GetPluginPath(ts.T(), "1.0.1-alpha1")

	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	ts.Run("try install plugin with alpha1 prerelease - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "install", "--path", pluginPath, "github.com/hashicorp/tester").
			Assert(check.MustFail(), check.Grep("Packer can only install plugin releases with this command", check.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestRemoteInstallWithPluginsInstall() {
	ts.SkipNoAcc()

	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	ts.Run("install latest version of a remote plugin with packer plugins install", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups").
			Assert(check.MustSucceed())
	})
}

func (ts *PackerPluginTestSuite) TestRemoteInstallOfPreReleasePlugin() {
	ts.SkipNoAcc()

	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	ts.Run("try to init with a pre-release constraint - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "templates/pre-release_constraint.pkr.hcl").
			Assert(check.MustFail(),
				check.Grep("Invalid version constraint", check.GrepStdout),
				check.Grep("Unsupported prerelease for constraint", check.GrepStdout))
	})

	ts.Run("try to plugins install with a pre-release version - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugin", "install", "github.com/hashicorp/hashicups", "1.0.2-dev").
			Assert(check.MustFail(),
				check.Grep("Unsupported prerelease for constraint", check.GrepStdout))
	})
}
