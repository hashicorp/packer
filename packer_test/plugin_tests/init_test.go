// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package plugin_tests

import (
	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerPluginTestSuite) TestPackerInitForce() {
	ts.SkipNoAcc()

	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	ts.Run("installs any missing plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", check.GrepStdout))
	})

	ts.Run("reinstalls plugins matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--force", "./templates/init/hashicups.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", check.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitUpgrade() {
	ts.SkipNoAcc()

	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	cmd := ts.PackerCommand().UsePluginDir(pluginPath)
	cmd.SetArgs("plugins", "install", "github.com/hashicorp/hashicups", "1.0.1")
	cmd.SetAssertFatal()
	cmd.Assert(check.MustSucceed(), check.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.1", check.GrepStdout))

	ts.Run("upgrades a plugin to the latest matching version constraints", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "--upgrade", "./templates/init/hashicups.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("Installed plugin github.com/hashicorp/hashicups v1.0.2", check.GrepStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitWithNonGithubSource() {
	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	ts.Run("try installing from an unreachable non-github source, should fail", func() {
		// The .invalid TLD never resolves (RFC 6761), so the remote getter
		// fails deterministically without depending on a live external host.
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(check.MustFail(), check.Grep(`failed to fetch https://hubgit.invalid/hashicorp/packer-plugin-tester/index.json`, check.GrepStdout))
	})

	ts.Run("manually install plugin to the expected source", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.10"), "hubgit.invalid/hashicorp/tester").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1.0.10", check.GrepStdout))
	})

	ts.Run("re-run packer init on same template, should succeed silently", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/non_gh.pkr.hcl").
			Assert(check.MustSucceed(),
				check.MkPipeCheck("no output in stdout").SetTester(check.ExpectEmptyInput()).SetStream(check.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestPackerInitWithMixedVersions() {
	ts.SkipNoAcc()

	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	ts.Run("skips the plugin installation with mixed versions before exiting with an error", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/init/mixed_versions.pkr.hcl").
			Assert(check.MustFail(),
				check.Grep("binary reported a pre-release version of 10.7.3-dev", check.GrepStdout))
	})
}
