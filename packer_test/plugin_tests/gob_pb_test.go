package plugin_tests

import "github.com/hashicorp/packer/packer_test/common/check"

const pbPluginName = "github.com/hashicorp/pbtester"

func CheckPBUsed(expect bool) check.Checker {
	const strToLookFor = "protobuf for communication with plugins"

	var opts []check.GrepOpts
	if !expect {
		opts = append(opts, check.GrepInvert)
	}

	return check.Grep(strToLookFor, opts...)
}

// Two different plugins installed locally, one with gob, one with protobuf.
// Both should have different sources so Packer will discover and fallback to using only gob.
func (ts *PackerGobTestSuite) TestTwoPluginsDifferentPB() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+gob")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0+pb"), pbPluginName).
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_both_plugins.pkr.hcl").
		Assert(CheckPBUsed(false))

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(CheckPBUsed(false))
}

// Two plugins, both with protobuf supported
// Both installed plugins will support protobuf, so Packer will use Protobuf for all its communications.
func (ts *PackerGobTestSuite) TestTwoPluginsBothPB() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0+pb"), pbPluginName).
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_both_plugins.pkr.hcl").
		Assert(CheckPBUsed(true))

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(CheckPBUsed(true))
}

// Two plugins, both with protobuf supported, force gob
// Both installed plugins support protobuf, but the environment variable PACKER_FORCE_GOB is
// set to 1 (or on), so Packer must use gob despite protobuf being supported all around.
func (ts *PackerGobTestSuite) TestTwoPluginsBothPBForceGob() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.0.0+pb"), pbPluginName).
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		AddEnv("PACKER_FORCE_GOB", "1").
		SetArgs("build", "./templates/test_both_plugins.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))

	ts.PackerCommand().UsePluginDir(pluginDir).
		AddEnv("PACKER_FORCE_GOB", "1").
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// Two plugins installed, one with two versions: one version supporting pb,
// one older with gob only. The other with only protobuf.
// The template used pins the older version of the first plugin.
// In this case, gob should be the one used, as the selected version supports
// gob only, despite a newer version supporting protobuf, and the other plugin
// also being compatible.
func (ts *PackerGobTestSuite) TestTwoPluginsLatestPBOlderGob_OlderPinned() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+gob", "1.1.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.GetPluginPath(ts.T(), "1.1.0+pb"), pbPluginName).
		Assert(check.MustSucceed(), check.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_pinned_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// One plugin installed, one version supporting pb, one older with gob only
// The template used pins the older version.
// In this case, gob should be the one used, as the selected version supports
// gob only, despite a newer version supporting protobuf.
func (ts *PackerGobTestSuite) TestOnePluginLatestPBOlderGob_OlderPinned() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+gob", "1.1.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_pinned_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// One plugin, with latest version supporting gob, but the older supporting protobuf
// In this case, Packer will default to using the latest version, and should
// default to using gob.
func (ts *PackerGobTestSuite) TestOnePluginWithLatestOnlyGob() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+pb", "1.1.0+gob")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// One plugin, gob only supported
// Packer will load the only plugin available there, and will use it, and use gob for comms
func (ts PackerGobTestSuite) TestOnePluginWithOnlyGob() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+gob")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// One plugin, protobuf supported
// Packer will load the only plugin available there, and use protobuf for comms
func (ts PackerGobTestSuite) TestOnePluginWithPB() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(true))
}

// No plugin installed, only internal components
// In this test, Packer must use Protobuf for internal components as nothing installed will prevent it.
func (ts PackerGobTestSuite) TestInternalOnly() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions()
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(true))
}

// One plugin with gob only installed, use only internal components
//
// Packer in this case will fallback to Gob, even if the template uses internal
// components only, as this is determined at loading time.
func (ts PackerGobTestSuite) TestInternalOnlyWithGobPluginInstalled() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+gob")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(false))
}

// One plugin with pb support installed, use only internal components
//
// Packer in this case will fallback to Gob, even if the template uses internal
// components only, as this is determined at loading time.
func (ts PackerGobTestSuite) TestInternalOnlyWithPBPluginInstalled() {
	pluginDir := ts.MakePluginDir().InstallPluginVersions("1.0.0+pb")
	defer pluginDir.Cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(check.MustSucceed(), CheckPBUsed(true))
}
