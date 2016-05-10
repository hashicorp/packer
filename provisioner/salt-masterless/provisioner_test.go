package saltmasterless

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"local_state_tree": os.TempDir(),
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.TempConfigDir != DefaultTempConfigDir {
		t.Errorf("unexpected temp config dir: %s", p.config.TempConfigDir)
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_MinionConfig(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["minion_config"] = "/i/dont/exist/i/think"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	tf, err := ioutil.TempFile("", "minion")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["minion_config"] = tf.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_MinionConfig_RemoteStateTree(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["minion_config"] = "/i/dont/exist/i/think"
	config["remote_state_tree"] = "/i/dont/exist/remote_state_tree"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("minion_config and remote_state_tree should cause error")
	}
}

func TestProvisionerPrepare_MinionConfig_RemotePillarRoots(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["minion_config"] = "/i/dont/exist/i/think"
	config["remote_pillar_roots"] = "/i/dont/exist/remote_pillar_roots"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("minion_config and remote_pillar_roots should cause error")
	}
}

func TestProvisionerPrepare_LocalStateTree(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["local_state_tree"] = "/i/dont/exist/i/think"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["local_state_tree"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_LocalPillarRoots(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["local_pillar_roots"] = "/i/dont/exist/i/think"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["local_pillar_roots"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerSudo(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	withSudo := p.sudo("echo hello")
	if withSudo != "sudo echo hello" {
		t.Fatalf("sudo command not generated correctly")
	}

	config["disable_sudo"] = true
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	withoutSudo := p.sudo("echo hello")
	if withoutSudo != "echo hello" {
		t.Fatalf("sudo-less command not generated correctly")
	}
}

func TestProvisionerPrepare_RemoteStateTree(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["remote_state_tree"] = "/remote_state_tree"
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "--file-root=/remote_state_tree") {
		t.Fatal("--file-root should be set in CmdArgs")
	}
}

func TestProvisionerPrepare_RemotePillarRoots(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["remote_pillar_roots"] = "/remote_pillar_roots"
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "--pillar-root=/remote_pillar_roots") {
		t.Fatal("--pillar-root should be set in CmdArgs")
	}
}

func TestProvisionerPrepare_RemoteStateTree_Default(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// no minion_config, no remote_state_tree
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "--file-root=/srv/salt") {
		t.Fatal("--file-root should be set in CmdArgs")
	}
}

func TestProvisionerPrepare_RemotePillarRoots_Default(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// no minion_config, no remote_pillar_roots
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "--pillar-root=/srv/pillar") {
		t.Fatal("--pillar-root should be set in CmdArgs")
	}
}

func TestProvisionerPrepare_NoExitOnFailure(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "--retcode-passthrough") {
		t.Fatal("--retcode-passthrough should be set in CmdArgs")
	}

	config["no_exit_on_failure"] = true
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if strings.Contains(p.config.CmdArgs, "--retcode-passthrough") {
		t.Fatal("--retcode-passthrough should not be set in CmdArgs")
	}
}

func TestProvisionerPrepare_LogLevel(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "-l info") {
		t.Fatal("-l info should be set in CmdArgs")
	}

	config["log_level"] = "debug"
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(p.config.CmdArgs, "-l debug") {
		t.Fatal("-l debug should be set in CmdArgs")
	}
}
