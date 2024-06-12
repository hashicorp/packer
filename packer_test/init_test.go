package packer_test

func (ts *PackerTestSuite) TestPackerInitForce() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("installs any missing plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(MustSucceed(), Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", grepStdout))
	})

	ts.Run("reinstalls plugins matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(MustSucceed(), Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", grepStdout))
	})
}

func (ts *PackerTestSuite) TestPackerInitUpgrade() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	cmd := ts.PackerCommand().UsePluginDir(pluginPath)
	cmd.SetArgs("plugins", "install", "github.com/hashicorp/hashicups", "1.0.1")
	cmd.Assert(MustSucceed(), Grep("Installed plugin github.com/hashicorp/hashicups v1.0.1", grepStdout))

	_, _, err := cmd.Run()
	if err != nil {
		ts.T().Fatalf("packer plugins install failed to install previous version of hashicups: %q", err)
	}

	ts.Run("upgrades a plugin to the latest matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--upgrade", "./templates/init/hashicups.pkr.hcl").
			Assert(MustSucceed(), Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", grepStdout))
	})
}

func (ts *PackerTestSuite) TestPackerInitWithNonGithubSource() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try installing from a non-github source, should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(MustFail(), Grep(`doesn't appear to be a valid "github.com" source address`, grepStdout))
	})

	ts.Run("manually install plugin to the expected source", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "--path", BuildSimplePlugin("1.0.10", ts.T()), "hubgit.com/hashicorp/tester").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.10", grepStdout))
	})

	ts.Run("re-run packer init on same template, should succeed silently", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(MustSucceed(),
				MkPipeCheck("no output in stdout").SetTester(ExpectEmptyInput()).SetStream(OnlyStdout))
	})
}

func (ts *PackerTestSuite) TestPackerInitWithMixedVersions() {
	ts.SkipNoAcc()

	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("skips the plugin installation with mixed versions before exiting with an error", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/mixed_versions.pkr.hcl").
			Assert(MustFail(),
				Grep("binary reported a pre-release version of 10.7.3-dev", grepStdout))
	})
}
