package packer_test

import (
	"strings"
)

func (ts *PackerTestSuite) TestPluginsRemoveWithSourceAddress() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10", "2.0.0")
	defer cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address removes all installed plugin versions", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester").
			Assert(MustSucceed(),
				Grep("packer-plugin-tester_v1.0.9", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepStdout),
				Grep("packer-plugin-tester_v2.0.0", grepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 0 {
		ts.T().Fatalf("Expected there to be 0 installed plugins but we got  %v", n)
	}
}

func (ts *PackerTestSuite) TestPluginsRemoveWithSourceAddressAndVersion() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.9", "1.0.10", "2.0.0")
	defer cleanup()

	// Get installed plugins
	if n := InstalledPlugins(ts, pluginPath); len(n) != 3 {
		ts.T().Fatalf("Expected there to be 3 installed plugins but we got  %v", n)
	}

	ts.Run("plugins remove with source address and version removes only the versioned plugin", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "remove", "github.com/hashicorp/tester", ">= 2.0.0").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v2.0.0", grepStdout))
	})

	ts.Run("plugins installed after single plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				MustSucceed(),
				Grep("packer-plugin-tester_v1.0.9", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepStdout),
				Grep("packer-plugin-tester_v2.0.0", grepInvert, grepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 2 {
		ts.T().Fatalf("Expected there to be 2 installed plugins but we got  %v", n)
	}
}

func (ts *PackerTestSuite) TestPluginsRemoveWithLocalPath() {
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
				MustSucceed(),
				Grep("packer-plugin-tester_v1.0.9", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepInvert, grepStdout),
			)
	})
	ts.Run("plugins installed after calling plugins remove outputs remaining installed plugins", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(
				MustSucceed(),
				Grep("packer-plugin-tester_v1.0.10", grepStdout),
				Grep("packer-plugin-tester_v1.0.9", grepInvert, grepStdout),
			)
	})

	// Get installed plugins after removal
	if n := InstalledPlugins(ts, pluginPath); len(n) != 1 {
		ts.T().Fatalf("Expected there to be 1 installed plugins but we got  %v", n)
	}
}

func InstalledPlugins(ts *PackerTestSuite, dir string) []string {
	ts.T().Helper()

	cmd := ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "installed")
	cmd.Assert(MustSucceed())
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
