package plugin_tests

import "github.com/hashicorp/packer/packer_test/lib"

func (ts *PackerPluginTestSuite) TestPackerInitForce() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("installs any missing plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", lib.GrepStdout))
	})

	ts.Run("reinstalls plugins matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", lib.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitUpgrade() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	cmd := ts.PackerCommand().UsePluginDir(pluginPath)
	cmd.SetArgs("plugins", "install", "github.com/hashicorp/hashicups", "1.0.1")
	cmd.Assert(lib.MustSucceed(), lib.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.1", lib.GrepStdout))

	_, _, err := cmd.Run()
	if err != nil {
		ts.T().Fatalf("packer plugins install failed to install previous version of hashicups: %q", err)
	}

	ts.Run("upgrades a plugin to the latest matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--upgrade", "./templates/init/hashicups.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", lib.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitWithNonGithubSource() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try installing from a non-github source, should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(lib.MustFail(), lib.Grep(`doesn't appear to be a valid "github.com" source address`, lib.GrepStdout))
	})

	ts.Run("manually install plugin to the expected source", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.10", ts.T()), "hubgit.com/hashicorp/tester").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepStdout))
	})

	ts.Run("re-run packer init on same template, should succeed silently", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(lib.MustSucceed(),
				lib.MkPipeCheck("no output in stdout").SetTester(lib.ExpectEmptyInput()).SetStream(lib.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitWithMixedVersions() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("skips the plugin installation with mixed versions before exiting with an error", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/mixed_versions.pkr.hcl").
			Assert(lib.MustFail(),
				lib.Grep("binary reported a pre-release version of 10.7.3-dev", lib.GrepStdout))
	})
}
