package plugin_tests

import "github.com/hashicorp/packer/packer_test/lib"

func (ts *PackerPluginTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install plugin with metadata in version", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0+metadata", ts.T()), "github.com/hashicorp/tester").
			Assert(lib.MustSucceed(), lib.Grep("Successfully installed plugin", lib.GrepStdout))
	})

	ts.Run("metadata plugin installed must not have metadata in its path", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1.0.0[^+]", lib.GrepStdout))
	})

	ts.Run("plugin with metadata should work with validate", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("validate", "./templates/simple.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", lib.GrepStderr))
	})

	ts.Run("plugin with metadata should work with build", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("build", "./templates/simple.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", lib.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestInstallPluginWithPath() {
	tempPluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install plugin with pre-release only", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0-dev", ts.T()), "github.com/hashicorp/tester").
			Assert(lib.MustSucceed(), lib.Grep("Successfully installed plugin", lib.GrepStdout))
	})

	ts.Run("install same plugin with pre-release + metadata", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0-dev+metadata", ts.T()), "github.com/hashicorp/tester").
			Assert(lib.MustSucceed(), lib.Grep("Successfully installed plugin", lib.GrepStdout))
	})

	ts.Run("list plugins, should only report one plugin", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(),
				lib.Grep("plugin-tester_v1.0.0-dev[^+]", lib.GrepStdout),
				lib.Grep("plugin-tester_v1.0.0-dev\\+", lib.GrepStdout, lib.GrepInvert),
				lib.LineCountCheck(1))
	})
}

func (ts *PackerPluginTestSuite) TestInstallPluginPrerelease() {
	pluginPath := ts.BuildSimplePlugin("1.0.1-alpha1", ts.T())

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try install plugin with alpha1 prerelease - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "install", "--path", pluginPath, "github.com/hashicorp/tester").
			Assert(lib.MustFail(), lib.Grep("Packer can only install plugin releases with this command", lib.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestRemoteInstallWithPluginsInstall() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install latest version of a remote plugin with packer plugins install", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups").
			Assert(lib.MustSucceed())
	})
}

func (ts *PackerPluginTestSuite) TestRemoteInstallOfPreReleasePlugin() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try to init with a pre-release constraint - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "templates/pre-release_constraint.pkr.hcl").
			Assert(lib.MustFail(),
				lib.Grep("Invalid version constraint", lib.GrepStdout),
				lib.Grep("Unsupported prerelease for constraint", lib.GrepStdout))
	})

	ts.Run("try to plugins install with a pre-release version - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugin", "install", "github.com/hashicorp/hashicups", "1.0.2-dev").
			Assert(lib.MustFail(),
				lib.Grep("Unsupported prerelease for constraint", lib.GrepStdout))
	})
}
