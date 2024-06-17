package packer_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func (ts *PackerTestSuite) TestLoadingOrder() {
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
					Assert(MustSucceed(), Grep(tt.grepStr))
			})
		}
	}
}

func (ts *PackerTestSuite) TestLoadWithLegacyPluginName() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	plugin := BuildSimplePlugin("1.0.10", ts.T())

	CopyFile(ts.T(), filepath.Join(pluginDir, "packer-plugin-tester"), plugin)

	ts.Run("only legacy plugins installed: expect build to fail", func() {
		ts.Run("with required_plugins - expect prompt for packer init", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.pkr.hcl").
				Assert(MustFail(),
					Grep("Did you run packer init for this project", grepStdout),
					Grep("following plugins are required", grepStdout))
		})

		ts.Run("JSON template, without required_plugins: should say the component is unknown", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.json").
				Assert(MustFail(),
					Grep("The builder tester-dynamic is unknown by Packer", grepStdout))
		})
	})

	pluginDir, cleanup = ts.MakePluginDir("1.0.0")
	defer cleanup()

	CopyFile(ts.T(), filepath.Join(pluginDir, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: one with no version in path, one with qualified name. Should pick-up the qualified one only.", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("build", "templates/simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", grepStderr))
	})

	wd, cleanup := TempWorkdir(ts.T(), "./templates/simple.pkr.hcl")
	defer cleanup()

	CopyFile(ts.T(), filepath.Join(wd, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: 1.0.0 in plugin dir with sum, one in workdir (no version). Should load 1.0.0", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).SetWD(wd).
			SetArgs("build", "simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", grepStderr))
	})
}

func (ts *PackerTestSuite) TestLoadWithSHAMismatches() {
	plugin := BuildSimplePlugin("1.0.10", ts.T())

	ts.Run("move plugin with right name, but no SHA256SUM, should reject", func() {
		pluginDir, cleanup := ts.MakePluginDir("1.0.9")
		defer cleanup()

		pluginDestName := ExpectedInstalledName("1.0.10")

		CopyFile(ts.T(), filepath.Join(pluginDir, "github.com", "hashicorp", "tester", pluginDestName), plugin)

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(),
				Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepStdout, grepInvert),
				Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", grepStderr))
	})

	ts.Run("move plugin with right name, invalid SHA256SUM, should reject", func() {
		pluginDir, cleanup := ts.MakePluginDir("1.0.9")
		defer cleanup()

		pluginDestName := ExpectedInstalledName("1.0.10")
		CopyFile(ts.T(), filepath.Join(pluginDir, "github.com", "hashicorp", "tester", pluginDestName), plugin)
		WriteFile(ts.T(),
			filepath.Join(pluginDir, "github.com", "hashicorp", "tester", fmt.Sprintf("%s_SHA256SUM", pluginDestName)),
			fmt.Sprintf("%x", sha256.New().Sum([]byte("Not the plugin's contents for sure."))))

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(),
				Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepInvert, grepStdout),
				Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", grepStderr),
				Grep(`Checksums \(\*sha256\.digest\) did not match.`, grepStderr))
	})
}

func (ts *PackerTestSuite) TestPluginPathEnvvarWithMultiplePaths() {
	pluginDirOne, cleanup := ts.MakePluginDir("1.0.10")
	defer cleanup()

	pluginDirTwo, cleanup := ts.MakePluginDir("1.0.9")
	defer cleanup()

	pluginDirVal := fmt.Sprintf("%s%c%s", pluginDirOne, os.PathListSeparator, pluginDirTwo)
	ts.Run("load plugin with two dirs - not supported anymore, should error", func() {
		ts.PackerCommand().UsePluginDir(pluginDirVal).
			SetArgs("plugins", "installed").
			Assert(MustFail(),
				Grep("Multiple paths are no longer supported for PACKER_PLUGIN_PATH"),
				MkPipeCheck("All envvars are suggested",
					PipeGrep(`\* PACKER_PLUGIN_PATH=`),
					LineCount()).
					SetStream(OnlyStderr).
					SetTester(IntCompare(eq, 2)))
	})
}

func (ts *PackerTestSuite) TestInstallNonCanonicalPluginVersion() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath, "github.com", "hashicorp", "tester"),
		BuildSimplePlugin("1.0.10", ts.T()),
		"001.00.010")

	ts.Run("try listing plugins with non-canonical version installed - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(),
				Grep(`version .* in path is non canonical`, grepStderr),
				MkPipeCheck("no output in stdout").SetTester(ExpectEmptyInput()).SetStream(OnlyStdout))
	})
}

func (ts *PackerTestSuite) TestLoadPluginWithMetadataInName() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ManualPluginInstall(ts.T(),
		filepath.Join(pluginPath, "github.com", "hashicorp", "tester"),
		BuildSimplePlugin("1.0.10+metadata", ts.T()),
		"1.0.10+metadata")

	ts.Run("try listing plugins with metadata in name - report none", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(),
				Grep("found version .* with metadata in the name", grepStderr),
				MkPipeCheck("no output in stdout").SetTester(ExpectEmptyInput()).SetStream(OnlyStdout))
	})
}

func (ts *PackerTestSuite) TestLoadWithOnlyReleaseFlag() {
	pluginPath, cleanup := ts.MakePluginDir("1.0.0", "1.0.1-dev")
	defer cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("run %s without --ignore-prerelease flag - pick 1.0.1-dev by default", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "./templates/simple.pkr.hcl").
				Assert(MustSucceed(),
					Grep("packer-plugin-tester_v1.0.1-dev.*: plugin process exited", grepStderr))
		})

		ts.Run(fmt.Sprintf("run %s with --ignore-prerelease flag - pick 1.0.0", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginPath).
				SetArgs(cmd, "--ignore-prerelease-plugins", "./templates/simple.pkr.hcl").
				Assert(MustSucceed(),
					Grep("packer-plugin-tester_v1.0.0.*: plugin process exited", grepStderr))
		})
	}
}

func (ts *PackerTestSuite) TestWithLegacyConfigAndComponents() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0")
	defer cleanup()

	workdir, cleanup := TempWorkdir(ts.T(), "./sample_config.json", "./templates/simple.json", "./templates/simple.pkr.hcl")
	defer cleanup()

	for _, cmd := range []string{"validate", "build"} {
		ts.Run(fmt.Sprintf("%s simple JSON template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.json").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(MustFail(),
					Grep("Your configuration file describes some legacy components", grepStderr),
					Grep("packer-provisioner-super-shell", grepStderr))
		})

		ts.Run(fmt.Sprintf("%s simple HCL2 template with config.json and components defined", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).SetWD(workdir).
				SetArgs(cmd, "simple.pkr.hcl").
				AddEnv("PACKER_CONFIG", filepath.Join(workdir, "sample_config.json")).
				Assert(MustFail(),
					Grep("Your configuration file describes some legacy components", grepStderr),
					Grep("packer-provisioner-super-shell", grepStderr))
		})
	}
}
