package gob_test

import (
	"github.com/hashicorp/packer/packer_test/lib"
)

const pbPluginName = "github.com/hashicorp/pbtester"

func CheckPBUsed(expect bool) lib.Checker {
	const strToLookFor = "protobuf for communication with plugins"

	var opts []lib.GrepOpts
	if !expect {
		opts = append(opts, lib.GrepInvert)
	}

	return lib.Grep(strToLookFor, opts...)
}

// Two different plugins installed locally, one with gob, one with protobuf.
// Both should have different sources so Packer will discover and fallback to using only gob.
func (ts *PackerGobTestSuite) TestTwoPluginsDifferentPB() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+gob")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0+pb", ts.T()), pbPluginName).
		Assert(lib.MustSucceed())

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
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0+pb", ts.T()), pbPluginName).
		Assert(lib.MustSucceed())

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
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.0.0+pb", ts.T()), pbPluginName).
		Assert(lib.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		AddEnv("PACKER_FORCE_GOB", "1").
		SetArgs("build", "./templates/test_both_plugins.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))

	ts.PackerCommand().UsePluginDir(pluginDir).
		AddEnv("PACKER_FORCE_GOB", "1").
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// Two plugins installed, one with two versions: one version supporting pb,
// one older with gob only. The other with only protobuf.
// The template used pins the older version of the first plugin.
// In this case, gob should be the one used, as the selected version supports
// gob only, despite a newer version supporting protobuf, and the other plugin
// also being compatible.
func (ts *PackerGobTestSuite) TestTwoPluginsLatestPBOlderGob_OlderPinned() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+gob", "1.1.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("plugins", "install", "--path", ts.BuildSimplePlugin("1.1.0+pb", ts.T()), pbPluginName).
		Assert(lib.MustSucceed(), lib.MustSucceed())

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_pinned_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// One plugin installed, one version supporting pb, one older with gob only
// The template used pins the older version.
// In this case, gob should be the one used, as the selected version supports
// gob only, despite a newer version supporting protobuf.
func (ts *PackerGobTestSuite) TestOnePluginLatestPBOlderGob_OlderPinned() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+gob", "1.1.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_pinned_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// One plugin, with latest version supporting gob, but the older supporting protobuf
// In this case, Packer will default to using the latest version, and should
// default to using gob.
func (ts *PackerGobTestSuite) TestOnePluginWithLatestOnlyGob() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+pb", "1.1.0+gob")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// One plugin, gob only supported
// Packer will load the only plugin available there, and will use it, and use gob for comms
func (ts PackerGobTestSuite) TestOnePluginWithOnlyGob() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+gob")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// One plugin, protobuf supported
// Packer will load the only plugin available there, and use protobuf for comms
func (ts PackerGobTestSuite) TestOnePluginWithPB() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/test_one_plugin.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(true))
}

// No plugin installed, only internal components
// In this test, Packer must use Protobuf for internal components as nothing installed will prevent it.
func (ts PackerGobTestSuite) TestInternalOnly() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(true))
}

// One plugin with gob only installed, use only internal components
//
// Packer in this case will fallback to Gob, even if the template uses internal
// components only, as this is determined at loading time.
func (ts PackerGobTestSuite) TestInternalOnlyWithGobPluginInstalled() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+gob")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(false))
}

// One plugin with pb support installed, use only internal components
//
// Packer in this case will fallback to Gob, even if the template uses internal
// components only, as this is determined at loading time.
func (ts PackerGobTestSuite) TestInternalOnlyWithPBPluginInstalled() {
	pluginDir, cleanup := ts.MakePluginDir("1.0.0+pb")
	defer cleanup()

	ts.PackerCommand().UsePluginDir(pluginDir).
		SetArgs("build", "./templates/internal_only.pkr.hcl").
		Assert(lib.MustSucceed(), CheckPBUsed(true))
}
