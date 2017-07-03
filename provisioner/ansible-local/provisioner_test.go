package ansiblelocal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmt"
	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	m := make(map[string]interface{})
	return m
}

func createTempFile() string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		panic(fmt.Sprintf("err: %s", err))
	}
	return file.Name()
}

func createTempFiles(numFiles int) []string {
	files := make([]string, 0, numFiles)
	defer func() {
		// Cleanup the files if not all were created.
		if len(files) < numFiles {
			for _, file := range files {
				os.Remove(file)
			}
		}
	}()

	for i := 0; i < numFiles; i++ {
		files = append(files, createTempFile())
	}
	return files
}

func removeFiles(files ...string) {
	for _, file := range files {
		os.Remove(file)
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

	if !strings.HasPrefix(filepath.ToSlash(p.config.StagingDir), DefaultStagingDir) {
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

func TestProvisionerPrepare_PlaybookFiles(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["playbook_file"] = ""
	config["playbook_files"] = []string{}
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
	config["playbook_files"] = []string{"some_other_file"}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["playbook_file"] = playbook_file.Name()
	config["playbook_files"] = []string{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config["playbook_file"] = ""
	config["playbook_files"] = []string{playbook_file.Name()}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func assertPlaybooksExecuted(comm *communicatorMock, playbooks []string) {
	cmdIndex := 0
	for _, playbook := range playbooks {
		for ; cmdIndex < len(comm.startCommand); cmdIndex++ {
			cmd := comm.startCommand[cmdIndex]
			if strings.Contains(cmd, "ansible-playbook") && strings.Contains(cmd, playbook) {
				break
			}
		}
		if cmdIndex == len(comm.startCommand) {
			panic(fmt.Sprintf("Playbook %s was not executed", playbook))
		}
	}
}

func assertPlaybooksUploaded(comm *communicatorMock, playbooks []string) {
	uploadIndex := 0
	for _, playbook := range playbooks {
		for ; uploadIndex < len(comm.uploadDestination); uploadIndex++ {
			dest := comm.uploadDestination[uploadIndex]
			if strings.HasSuffix(dest, playbook) {
				break
			}
		}
		if uploadIndex == len(comm.uploadDestination) {
			panic(fmt.Sprintf("Playbook %s was not uploaded", playbook))
		}
	}
}

func TestProvisionerProvision_PlaybookFiles(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbooks := createTempFiles(3)
	defer removeFiles(playbooks...)

	config["playbook_files"] = playbooks
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	if err := p.Provision(&uiStub{}, comm); err != nil {
		t.Fatalf("err: %s", err)
	}

	assertPlaybooksUploaded(comm, playbooks)
	assertPlaybooksExecuted(comm, playbooks)
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
