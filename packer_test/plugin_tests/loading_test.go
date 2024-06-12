package plugin_tests

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer_test/lib"
)

func (ts *PackerPluginTestSuite) TestLoadingOrder() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.9", "1.0.10")
	defer cleanup()

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
					Assert(lib.MustSucceed(), lib.Grep(tt.grepStr))
			})
		}
	}
}

func (ts *PackerPluginTestSuite) TestLoadWithLegacyPluginName() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	plugin := ts.BuildSimplePlugin("1.0.10", ts.T())

	lib.CopyFile(ts.T(), filepath.Join(pluginDir, "packer-plugin-tester"), plugin)

	ts.Run("only legacy plugins installed: expect build to fail", func() {
		ts.Run("with required_plugins - expect prompt for packer init", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.pkr.hcl").
				Assert(lib.MustFail(),
					lib.Grep("Did you run packer init for this project", lib.GrepStdout),
					lib.Grep("following plugins are required", lib.GrepStdout))
		})

		ts.Run("JSON template, without required_plugins: should say the component is unknown", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.json").
				Assert(lib.MustFail(),
					lib.Grep("The builder tester-dynamic is unknown by Packer", lib.GrepStdout))
		})
	})

	pluginDir, cleanup = ts.MakePluginDir("1.0.0")
	defer cleanup()

	lib.CopyFile(ts.T(), filepath.Join(pluginDir, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: one with no version in path, one with qualified name. Should pick-up the qualified one only.", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("build", "templates/simple.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", lib.GrepStderr))
	})

	wd, cleanup := lib.TempWorkdir(ts.T(), "./templates/simple.pkr.hcl")
	defer cleanup()

	lib.CopyFile(ts.T(), filepath.Join(wd, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: 1.0.0 in plugin dir with sum, one in workdir (no version). Should load 1.0.0", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).SetWD(wd).
			SetArgs("build", "simple.pkr.hcl").
			Assert(lib.MustSucceed(), lib.Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", lib.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestLoadWithSHAMismatches() {
	plugin := ts.BuildSimplePlugin("1.0.10", ts.T())

	ts.Run("move plugin with right name, but no SHA256SUM, should reject", func() {
		pluginDir, cleanup := ts.MakePluginDir("1.0.9")
		defer cleanup()

		pluginDestName := lib.ExpectedInstalledName("1.0.10")

		lib.CopyFile(ts.T(), filepath.Join(pluginDir, "github.com", "hashicorp", "tester", pluginDestName), plugin)

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepStdout, lib.GrepInvert),
				lib.Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", lib.GrepStderr))
	})

	ts.Run("move plugin with right name, invalid SHA256SUM, should reject", func() {
		pluginDir, cleanup := ts.MakePluginDir("1.0.9")
		defer cleanup()

		pluginDestName := lib.ExpectedInstalledName("1.0.10")
		lib.CopyFile(ts.T(), filepath.Join(pluginDir, "github.com", "hashicorp", "tester", pluginDestName), plugin)
		lib.WriteFile(ts.T(),
			filepath.Join(pluginDir, "github.com", "hashicorp", "tester", fmt.Sprintf("%s_SHA256SUM", pluginDestName)),
			fmt.Sprintf("%x", sha256.New().Sum([]byte("Not the plugin's contents for sure."))))

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(),
				lib.Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", lib.GrepStdout),
				lib.Grep("packer-plugin-tester_v1.0.10", lib.GrepInvert, lib.GrepStdout),
				lib.Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", lib.GrepStderr),
				lib.Grep(`Checksums \(\*sha256\.digest\) did not match.`, lib.GrepStderr))
	})
}

func (ts *PackerPluginTestSuite) TestPluginPathEnvvarWithMultiplePaths() {
	pluginDirOne, cleanup := ts.MakePluginDir("1.0.10")
	defer cleanup()

	pluginDirTwo, cleanup := ts.MakePluginDir("1.0.9")
	defer cleanup()

	pluginDirVal := fmt.Sprintf("%s%c%s", pluginDirOne, os.PathListSeparator, pluginDirTwo)
	ts.Run("load plugin with two dirs - not supported anymore, should error", func() {
		ts.PackerCommand().UsePluginDir(pluginDirVal).
			SetArgs("plugins", "installed").
			Assert(lib.MustFail(),
				lib.Grep("Multiple paths are no longer supported for PACKER_PLUGIN_PATH"),
				lib.MkPipeCheck("All envvars are suggested",
					lib.PipeGrep(`\* PACKER_PLUGIN_PATH=`),
					lib.LineCount()).
					SetStream(lib.OnlyStderr).
					SetTester(lib.IntCompare(lib.Eq, 2)))
	})
}

func (ts *PackerPluginTestSuite) TestInstallNonCanonicalPluginVersion() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	lib.ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath, "github.com", "hashicorp", "tester"),
		ts.BuildSimplePlugin("1.0.10", ts.T()),
		"001.00.010")

	ts.Run("try listing plugins with non-canonical version installed - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(),
				lib.Grep(`version .* in path is non canonical`, lib.GrepStderr),
				lib.MkPipeCheck("no output in stdout").SetTester(lib.ExpectEmptyInput()).SetStream(lib.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestLoadPluginWithMetadataInName() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	lib.ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath, "github.com", "hashicorp", "tester"),
		ts.BuildSimplePlugin("1.0.10+metadata", ts.T()),
		"1.0.10+metadata")

	ts.Run("try listing plugins with metadata in name - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(lib.MustSucceed(),
				lib.Grep("found version .* with metadata in the name", lib.GrepStderr),
				lib.MkPipeCheck("no output in stdout").SetTester(lib.ExpectEmptyInput()).SetStream(lib.OnlyStdout))
	})
}

func (ts *PackerPluginTestSuite) TestLoadWithOnlyReleaseFlag() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.0", "1.0.1-dev")
	defer cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("run %s without --ignore-prerelease flag - pick 1.0.1-dev by default", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "./templates/simple.pkr.hcl").
				Assert(lib.MustSucceed(),
					lib.Grep("packer-plugin-tester_v1.0.1-dev.*: plugin process exited", lib.GrepStderr))
		})

		ts.Run(fmt.Sprintf("run %s with --ignore-prerelease flag - pick 1.0.0", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "--ignore-prerelease-plugins", "./templates/simple.pkr.hcl").
				Assert(lib.MustSucceed(),
					lib.Grep("packer-plugin-tester_v1.0.0.*: plugin process exited", lib.GrepStderr))
		})
	}
}

func (ts *PackerPluginTestSuite) TestWithLegacyConfigAndComponents() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0")
	defer cleanup()

	workdir, cleanup := lib.TempWorkdir(ts.T(), "./sample_config.json", "./templates/simple.json", "./templates/simple.pkr.hcl")
	defer cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("%s simple JSON template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.json").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(lib.MustFail(),
					lib.Grep("Your configuration file describes some legacy components", lib.GrepStderr),
					lib.Grep("packer-provisioner-super-shell", lib.GrepStderr))
		})

		ts.Run(fmt.Sprintf("%s simple HCL2 template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.pkr.hcl").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(lib.MustFail(),
					lib.Grep("Your configuration file describes some legacy components", lib.GrepStderr),
					lib.Grep("packer-provisioner-super-shell", lib.GrepStderr))
		})
	}
}
