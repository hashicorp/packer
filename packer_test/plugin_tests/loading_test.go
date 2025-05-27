package plugin_tests

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer_test/common"
	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerPluginTestSuite) TestLoadingOrder() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.9", "1.0.10")
	defer pluginDir.Cleanup()

	for _, command := range []string{"build", "validate"} {
		tests := []struct {
			name         string
			templatePath string
			grepStr      string
		}{
			{
				"HCL2 " + command + " - With required_plugins, 1.0.10 is the most recent and should load",
				"./templates/simple.pkr.hcl",
				"packer-plugin-tester_v1\\.0\\.10[^\n]+ plugin:",
			},
			{
				"JSON " + command + " - No required_plugins, 1.0.10 is the most recent and should load",
				"./templates/simple.json",
				"packer-plugin-tester_v1\\.0\\.10[^\n]+ plugin:",
			},
			{
				"HCL2 " + command + " - With required_plugins, 1.0.9 is pinned, so 1.0.9 should be used",
				"./templates/pin_1.0.9.pkr.hcl",
				"packer-plugin-tester_v1\\.0\\.9[^\n]+ plugin:",
			},
		}

		for _, tt := range tests {
			ts.Run(tt.name, func() {
				ts.PackerCommand().
					SetArgs(command, tt.templatePath).
					UsePluginDir(pluginDir).
					Assert(check.MustSucceed(), check.Grep(tt.grepStr))
			})
		}
	}
}

func (ts *PackerPluginTestSuite) TestLoadWithLegacyPluginName() {
	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	plugin := ts.GetPluginPath(ts.T(), "1.0.10")

	common.CopyFile(ts.T(), filepath.Join(pluginDir.Dir(), "packer-plugin-tester"), plugin)

	ts.Run("only legacy plugins installed: expect build to fail", func() {
		ts.Run("with required_plugins - expect prompt for packer init", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.pkr.hcl").
				Assert(check.MustFail(),
					check.Grep("Did you run packer init for this project", check.GrepStdout),
					check.Grep("following plugins are required", check.GrepStdout))
		})

		ts.Run("JSON template, without required_plugins: should say the component is unknown", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.json").
				Assert(check.MustFail(),
					check.Grep("The builder tester-dynamic is unknown by Packer", check.GrepStdout))
		})
	})

	pluginDir = ts.MakePluginDir().InstallPluginVersions("1.0.0")
	defer pluginDir.Cleanup()

	common.CopyFile(ts.T(), filepath.Join(pluginDir.Dir(), "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: one with no version in path, one with qualified name. Should pick-up the qualified one only.", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("build", "templates/simple.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", check.GrepStderr))
	})

	wd, cleanup := common.TempWorkdir(ts.T(), "./templates/simple.pkr.hcl")
	defer cleanup()

	common.CopyFile(ts.T(), filepath.Join(wd, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: 1.0.0 in plugin dir with sum, one in workdir (no version). Should load 1.0.0", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).SetWD(wd).
			SetArgs("build", "simple.pkr.hcl").
			Assert(check.MustSucceed(), check.Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", check.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestLoadWithSHAMismatches() {
	plugin := ts.GetPluginPath(ts.T(), "1.0.10")

	ts.Run("move plugin with right name, but no SHA256SUM, should reject", func() {
		pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.9")
		defer pluginDir.Cleanup()

		pluginDestName := common.ExpectedInstalledName("1.0.10")

		common.CopyFile(ts.T(), filepath.Join(pluginDir.Dir(), "github.com", "hashicorp", "tester", pluginDestName), plugin)

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", check.GrepStdout),
				check.GrepInverted("packer-plugin-tester_v1.0.10", check.GrepStdout),
				check.Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", check.GrepStderr))
	})

	ts.Run("move plugin with right name, invalid SHA256SUM, should reject", func() {
		pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.9")
		defer pluginDir.Cleanup()

		pluginDestName := common.ExpectedInstalledName("1.0.10")
		common.CopyFile(ts.T(), filepath.Join(pluginDir.Dir(), "github.com", "hashicorp", "tester", pluginDestName), plugin)

		common.WriteFile(ts.T(),
			filepath.Join(pluginDir.Dir(), "github.com", "hashicorp", "tester", fmt.Sprintf("%s_SHA256SUM", pluginDestName)),
			fmt.Sprintf("%x", sha256.New().Sum([]byte("Not the plugin's contents for sure."))))

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(),
				check.Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", check.GrepStdout),
				check.GrepInverted("packer-plugin-tester_v1.0.10", check.GrepStdout),
				check.Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", check.GrepStderr),
				check.Grep(`Checksums \(\*sha256\.[dD]igest\) did not match.`, check.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestPluginPathEnvvarWithMultiplePaths() {
	pluginDirOne := ts.MakePluginDir().InstallPluginVersions("1.0.10")
	defer pluginDirOne.Cleanup()

	pluginDirTwo := ts.MakePluginDir().InstallPluginVersions("1.0.9")
	defer pluginDirTwo.Cleanup()

	pluginDirVal := fmt.Sprintf("%s%c%s", pluginDirOne.Dir(), os.PathListSeparator, pluginDirTwo.Dir())
	ts.Run("load plugin with two dirs - not supported anymore, should error", func() {
		ts.PackerCommand().UseRawPluginDir(pluginDirVal).
			SetArgs("plugins", "installed").
			Assert(check.MustFail(),
				check.Grep("Multiple paths are no longer supported for PACKER_PLUGIN_PATH"),
				check.MkPipeCheck("All envvars are suggested",
					check.PipeGrep(`\* PACKER_PLUGIN_PATH=`),
					check.LineCount()).
					SetStream(check.OnlyStderr).
					SetTester(check.IntCompare(check.Eq, 2)))
	})
}

func (ts *PackerPluginTestSuite) TestInstallNonCanonicalPluginVersion() {
	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	common.ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath.Dir(), "github.com", "hashicorp", "tester"),
		ts.GetPluginPath(ts.T(), "1.0.10"),
		"001.00.010")

	ts.Run("try listing plugins with non-canonical version installed - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(),
				check.Grep(`version .* in path is non canonical`, check.GrepStderr),
				check.MkPipeCheck("no output in stdout").SetTester(check.ExpectEmptyInput()).SetStream(check.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestLoadPluginWithMetadataInName() {
	pluginPath := ts.MakePluginDir()
	defer pluginPath.Cleanup()

	common.ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath.Dir(), "github.com", "hashicorp", "tester"),
		ts.GetPluginPath(ts.T(), "1.0.10+metadata"),
		"1.0.10+metadata")

	ts.Run("try listing plugins with metadata in name - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(check.MustSucceed(),
				check.Grep("found version .* with metadata in the name", check.GrepStderr),
				check.MkPipeCheck("no output in stdout").SetTester(check.ExpectEmptyInput()).SetStream(check.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestLoadWithOnlyReleaseFlag() {
	pluginPath := ts.MakePluginDir().InstallPluginVersions("1.0.0", "1.0.1-dev")
	defer pluginPath.Cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("run %s without --ignore-prerelease flag - pick 1.0.1-dev by default", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "./templates/simple.pkr.hcl").
				Assert(check.MustSucceed(),
					check.Grep("packer-plugin-tester_v1.0.1-dev.*: plugin process exited", check.GrepStderr))
		})

		ts.Run(fmt.Sprintf("run %s with --ignore-prerelease flag - pick 1.0.0", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "--ignore-prerelease-plugins", "./templates/simple.pkr.hcl").
				Assert(check.MustSucceed(),
					check.Grep("packer-plugin-tester_v1.0.0.*: plugin process exited", check.GrepStderr))
		})
	}
}

func (ts *PackerPluginTestSuite) TestWithLegacyConfigAndComponents() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0")
	defer pluginDir.Cleanup()

	workdir, cleanup := common.TempWorkdir(ts.T(), "./sample_config.json", "./templates/simple.json", "./templates/simple.pkr.hcl")
	defer cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("%s simple JSON template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.json").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(check.MustFail(),
					check.Grep("Your configuration file describes some legacy components", check.GrepStderr),
					check.Grep("packer-provisioner-super-shell", check.GrepStderr))
		})

		ts.Run(fmt.Sprintf("%s simple HCL2 template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.pkr.hcl").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(check.MustFail(),
					check.Grep("Your configuration file describes some legacy components", check.GrepStderr),
					check.Grep("packer-provisioner-super-shell", check.GrepStderr))
		})
	}
}
