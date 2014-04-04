package ansiblelocal

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	m := make(map[string]interface{})
	return m
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

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["playbook_file"] = playbook_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.StagingDir != DefaultStagingDir {
		t.Fatalf("unexpected staging dir %s, expected %s",
			p.config.StagingDir, DefaultStagingDir)
	}
}

func TestProvisionerPrepare_PlaybookFile(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["playbook_file"] = ""
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["playbook_file"] = playbook_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_InventoryFile(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["playbook_file"] = ""
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["playbook_file"] = playbook_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	inventory_file, err := ioutil.TempFile("", "inventory")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(inventory_file.Name())

	config["inventory_file"] = inventory_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_Dirs(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["playbook_file"] = ""
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["playbook_file"] = playbook_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config["playbook_paths"] = []string{playbook_file.Name()}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should error if playbook paths is not a dir")
	}

	config["playbook_paths"] = []string{os.TempDir()}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config["role_paths"] = []string{playbook_file.Name()}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should error if role paths is not a dir")
	}

	config["role_paths"] = []string{os.TempDir()}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config["group_vars"] = playbook_file.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should error if group_vars path is not a dir")
	}

	config["group_vars"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config["host_vars"] = playbook_file.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should error if host_vars path is not a dir")
	}

	config["host_vars"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}
