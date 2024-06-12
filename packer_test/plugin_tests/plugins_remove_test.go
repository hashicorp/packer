package plugin_tests

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer_test/lib"
)

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithSourceAddress() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10", "2.0.0")
	defer cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address removes all installed plugin versions", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester").
			Assert(lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1.0.9", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v2.0.0", lib.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 0 {
		ts.T().Fatalf("Expected there to be 0 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with incorrect source address exits non found error", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/testerONE").
			Assert(
				lib.MustFail(),
				lib.Grep("No installed plugin found matching the plugin constraints github.com/hashicorp/testerONE"),
			)
	})

	ts.Run("plugins remove with invalid source address exits with non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/a/a/a/a/a/a/a/a/a/a/a/a/a/a/a/tester").
			Assert(
				lib.MustFail(),
				lib.Grep("The source URL must have at most 16 components"),
			)
	})
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithSourceAddressAndVersion() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10", "2.0.0")
	defer cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address and version removes only the versioned plugin", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester", ">= 2.0.0").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v2.0.0", lib.GrepStdout))
	})

	ts.Run("plugins installed after single plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1.0.9", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v2.0.0", lib.GrepInvert, lib.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 2 {
		ts.T().Fatalf("Expected there to be 2 installed plugins but we got  %v", n)
	}
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithLocalPath() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10")
	defer cleanup()

	// Get installed plugins
	plugins := InstalledPlugins(ts, pluginPath)
	if len(plugins) != 2 {
		ts.T().Fatalf("Expected there to be 2 installed plugins but we got  %v", len(plugins))
	}

	ts.Run("plugins remove with a local path removes only the specified plugin", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", plugins[0]).
			Assert(
				lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1.0.9", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepInvert, lib.GrepStdout),
			)
	})
	ts.Run("plugins installed after calling plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.9", lib.GrepInvert, lib.GrepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with incomplete local path exits with a non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", filepath.Base(plugins[0])).
			Assert(
				lib.MustFail(),
				lib.Grep("A source URL must at least contain a host and a path with 2 components", lib.GrepStdout),
			)
	})

	ts.Run("plugins remove with fake local path exits with a non-zero code", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", ts.T().TempDir()).
			Assert(
				lib.MustFail(),
				lib.Grep("is not under the plugin directory inferred by Packer", lib.GrepStdout),
			)
	})
}

func (ts *PackerPluginTestSuite) TestPluginsRemoveWithNoArguments() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9")
	defer cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with no options returns non-zero with help text", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove").
			Assert(
				lib.MustFail(),
				lib.Grep("Usage: packer plugins remove <plugin>", lib.GrepStdout),
			)
	})

	// Get installed should remain the same
	if n := InstalledPlugins(ts, pluginPath); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}

}

func InstalledPlugins(ts *PackerPluginTestSuite, dir string) []string {
	ts.T().Helper()

	cmd := ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "installed")
	cmd.Assert(lib.MustSucceed())
	if ts.T().Failed() {
		ts.T().Fatalf("Failed to execute plugin installed for %q", dir)
	}

	out, _, _ := cmd.Run()
	// Output will be split on '\n' after trimming all other white space
	out = strings.TrimSpace(out)
	plugins := strings.Fields(out)
	n := len(plugins)
	if n == 0 {
		return nil
	}
	return plugins
}
