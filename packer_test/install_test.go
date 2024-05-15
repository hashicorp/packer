package packer_test

import "strings"

func (ts *PackerTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir, cleanup := ts.MakePluginDir("1.0.0+metadata")
	defer cleanup()

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
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install latest version of a remote plugin with packer plugins install", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups").
			Assert(MustSucceed())
	})

	ts.Run("install dev version of a remote plugin with packer plugins install - must fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups", "v1.0.2-dev").
			Assert(MustFail(), Grep("Remote installation of pre-release plugin versions is unsupported.", grepStdout))
	})
}

func (ts *PackerTestSuite) TestRemovePluginWithLocalPath() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10")
	defer cleanup()

	// Get installed plugins
	cmd := ts.PackerCommand().UsePluginDir(pluginPath).
		SetArgs("plugins", "installed")
	cmd.Assert(MustSucceed())
	if ts.T().Failed() {
		return
	}

	plugins, _, _ := cmd.Run()
	pluginList := strings.Split(strings.TrimSpace(plugins), "\n")
	if len(pluginList) != 2 {
		ts.T().Fatalf("Not the expected installed plugins: %v; expected 2", pluginList)
	}

	ts.Run("remove one plugin version with its local path", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", pluginList[0]).
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.9", grepStdout))
	})
	ts.Run("ensure one plugin remaining only", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				MustSucceed(),
				Grep("packer-plugin-tester_v1.0.10", grepStdout),
				Grep("packer-plugin-tester_v1.0.9", grepInvert, grepStdout),
			)
	})
	ts.Run("remove plugin with version", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester", "1.0.10").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.10", grepStdout))
	})
}
