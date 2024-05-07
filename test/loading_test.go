package test

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func (ts *PackerTestSuite) TestLoadingOrder() {
	t := ts.T()

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
			t.Run(tt.name, func(t *testing.T) {
				ts.PackerCommand().
					SetArgs(command, tt.templatePath).
					UsePluginDir(pluginDir).
					Assert(t, MustSucceed(), Grep(tt.grepStr))
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
				Assert(ts.T(), MustFail(),
					Grep("Did you run packer init for this project", grepStdout),
					Grep("following plugins are required", grepStdout))
		})

		ts.Run("JSON template, without required_plugins: should say the component is unknown", func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs("build", "templates/simple.json").
				Assert(ts.T(), MustFail(),
					Grep("The builder tester-dynamic is unknown by Packer", grepStdout))
		})
	})

	pluginDir, cleanup = ts.MakePluginDir("1.0.0")
	defer cleanup()

	CopyFile(ts.T(), filepath.Join(pluginDir, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: one with no version in path, one with qualified name. Should pick-up the qualified one only.", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("build", "templates/simple.pkr.hcl").
			Assert(ts.T(), MustSucceed(), Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", grepStderr))
	})

	wd, cleanup := TempWorkdir(ts.T(), "./templates/simple.pkr.hcl")
	defer cleanup()

	CopyFile(ts.T(), filepath.Join(wd, "packer-plugin-tester"), plugin)

	ts.Run("multiple plugins installed: 1.0.0 in plugin dir with sum, one in workdir (no version). Should load 1.0.0", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).SetWD(wd).
			SetArgs("build", "simple.pkr.hcl").
			Assert(ts.T(), MustSucceed(), Grep("packer-plugin-tester_v1\\.0\\.0[^\\n]+ plugin:", grepStderr))
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
			Assert(ts.T(), MustSucceed(),
				Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepStdout, grepInvert),
				Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", grepStderr))
	})

	ts.Run("move plugin with right name, invalid SHA256SUM, should reject", func() {
		pluginDir, cleanup := ts.MakePluginDir("1.0.9")
		defer cleanup()

		pluginDestName := ExpectedInstalledName("1.0.10")
		noExtDest := pluginDestName
		if runtime.GOOS == "windows" {
			noExtDest = strings.Replace(pluginDestName, ".exe", "", 1)
		}

		CopyFile(ts.T(), filepath.Join(pluginDir, "github.com", "hashicorp", "tester", pluginDestName), plugin)
		WriteFile(ts.T(),
			filepath.Join(pluginDir, "github.com", "hashicorp", "tester", fmt.Sprintf("%s_SHA256SUM", noExtDest)),
			fmt.Sprintf("%x", sha256.New().Sum([]byte("Not the plugin's contents for sure."))))

		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "installed").
			Assert(ts.T(), MustSucceed(),
				Grep("packer-plugin-tester_v1\\.0\\.9[^\\n]+", grepStdout),
				Grep("packer-plugin-tester_v1.0.10", grepInvert, grepStdout),
				Grep("v1.0.10[^\\n]+ignoring possibly unsafe binary", grepStderr),
				Grep(`Checksums \(\*sha256\.digest\) did not match.`, grepStderr))
	})
}
