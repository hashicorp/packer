//go:build linux

package plugin_tests

import (
	"os"

	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerShellProvisionerTestSuite) TestNoShebangInScript() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/no_shebang_in_script.pkr.hcl").
		Assert(check.MustSucceed())
}

func (ts *PackerShellProvisionerTestSuite) TestShebangInInlineScript() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/shebang_in_inline.pkr.hcl").
		Assert(check.MustSucceed())
}

func (ts *PackerShellProvisionerTestSuite) TestShebangAsOption() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/shebang_as_option.pkr.hcl").
		Assert(check.MustSucceed())
}

func (ts *PackerShellProvisionerTestSuite) TestShebangAsOptionNotInline() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/no_shebang_inline_but_as_option.pkr.hcl").
		Assert(check.MustSucceed())
}

func (ts *PackerShellProvisionerTestSuite) TestInvalidShebangAsOption() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/shebang_as_option_invalid.pkr.hcl").
		Assert(check.MustFail())
}

func (ts *PackerShellProvisionerTestSuite) TestEmptyInlineCommands() {
	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/empty_inline_list.pkr.hcl").
		Assert(check.MustFail())
}
