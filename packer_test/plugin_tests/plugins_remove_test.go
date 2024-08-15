package plugin_tests

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithSourceAddress() {
	pluginPath := ts.MakePluginDir().InstallPluginVersions("1.0.9", "1.0.10", "2.0.0")
	defer pluginPath.Cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address removes all installed plugin versions", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester").
			Assert(check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1.0.9", check.GrepStdout),
				check.Grep("packer-plugin-tester_v1.0.10", check.GrepStdout),
				check.Grep("packer-plugin-tester_v2.0.0", check.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 0 {
		ts.T().Fatalf("Expected there to be 0 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with incorrect source address exits non found error", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/testerONE").
			Assert(
				check.MustFail(),
				check.Grep("No installed plugin found matching the plugin constraints github.com/hashicorp/testerONE"),
			)
	})

	ts.Run("plugins remove with invalid source address exits with non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/a/a/a/a/a/a/a/a/a/a/a/a/a/a/a/tester").
			Assert(
				check.MustFail(),
				check.Grep("The source URL must have at most 16 components"),
			)
	})
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithSourceAddressAndVersion() {
	pluginPath := ts.MakePluginDir().InstallPluginVersions("1.0.9", "1.0.10", "2.0.0")
	defer pluginPath.Cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address and version removes only the versioned plugin", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester", ">= 2.0.0").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v2.0.0", check.GrepStdout))
	})

	ts.Run("plugins installed after single plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1.0.9", check.GrepStdout),
				check.Grep("packer-plugin-tester_v1.0.10", check.GrepStdout),
				check.GrepInverted("packer-plugin-tester_v2.0.0", check.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 2 {
		ts.T().Fatalf("Expected there to be 2 installed plugins but we got  %v", n)
	}
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithLocalPath() {
	pluginPath := ts.MakePluginDir().InstallPluginVersions("1.0.9", "1.0.10")
	defer pluginPath.Cleanup()

	// Get installed plugins
	plugins := InstalledPlugins(ts, pluginPath.Dir())
	if len(plugins) != 2 {
		ts.T().Fatalf("Expected there to be 2 installed plugins but we got  %v", len(plugins))
	}

	ts.Run("plugins remove with a local path removes only the specified plugin", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", plugins[0]).
			Assert(
				check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1.0.9", check.GrepStdout),
				check.GrepInverted("packer-plugin-tester_v1.0.10", check.GrepStdout),
			)
	})
	ts.Run("plugins installed after calling plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1.0.10", check.GrepStdout),
				check.GrepInverted("packer-plugin-tester_v1.0.9", check.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with incomplete local path exits with a non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", filepath.Base(plugins[0])).
			Assert(
				check.MustFail(),
				check.Grep("A source URL must at least contain a host and a path with 2 components", check.GrepStdout),
			)
	})

	ts.Run("plugins remove with fake local path exits with a non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", ts.T().TempDir()).
			Assert(
				check.MustFail(),
				check.Grep("is not under the plugin directory inferred by Packer", check.GrepStdout),
			)
	})
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithNoArguments() {
	pluginPath := ts.MakePluginDir().InstallPluginVersions("1.0.9")
	defer pluginPath.Cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with no options returns non-zero with help text", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove").
			Assert(
				check.MustFail(),
				check.Grep("Usage: packer plugins remove <plugin>", check.GrepStdout),
			)
	})

	// Get installed should remain the same
	if n := InstalledPlugins(ts, pluginPath.Dir()); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

}

func InstalledPlugins(ts *PackerPluginTestSuite, dir string) []string {
	ts.T().Helper()

	cmd := ts.PackerCommand().UseRawPluginDir(dir).
		SetArgs("plugins", "installed").
		SetAssertFatal()
	cmd.Assert(check.MustSucceed())

	out, _, _ := cmd.Output()
	// Output will be split on '\n' after trimming all other white space
	out = strings.TrimSpace(out)
	plugins := strings.Fields(out)
	n := len(plugins)
	if n == 0 {
		return nil
	}
	return plugins
}
