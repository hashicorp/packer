package ansiblelocal

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmt"
	"os/exec"

	"github.com/hashicorp/packer-plugin-docker/builder/docker"
	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/provisioner/file"
)

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
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

	p = Provisioner{}
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

func TestProvisionerProvision_PlaybookFiles(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbooks := createTempFiles("", 3)
	defer removeFiles(playbooks...)

	config["playbook_files"] = playbooks
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	if err := p.Provision(context.Background(), packersdk.TestUi(t), comm, make(map[string]interface{})); err != nil {
		t.Fatalf("err: %s", err)
	}

	assertPlaybooksUploaded(comm, playbooks)
	assertPlaybooksExecuted(comm, playbooks)
}

func TestProvisionerProvision_PlaybookFilesWithPlaybookDir(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbook_dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Failed to create playbook_dir: %s", err)
	}
	defer os.RemoveAll(playbook_dir)
	playbooks := createTempFiles(playbook_dir, 3)

	playbookNames := make([]string, 0, len(playbooks))
	playbooksInPlaybookDir := make([]string, 0, len(playbooks))
	for _, playbook := range playbooks {
		playbooksInPlaybookDir = append(playbooksInPlaybookDir, strings.TrimPrefix(playbook, playbook_dir))
		playbookNames = append(playbookNames, filepath.Base(playbook))
	}

	config["playbook_files"] = playbooks
	config["playbook_dir"] = playbook_dir
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	if err := p.Provision(context.Background(), packersdk.TestUi(t), comm, make(map[string]interface{})); err != nil {
		t.Fatalf("err: %s", err)
	}

	assertPlaybooksNotUploaded(comm, playbookNames)
	assertPlaybooksExecuted(comm, playbooksInPlaybookDir)
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

func TestProvisionerPrepare_CleanStagingDir(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["playbook_file"] = playbook_file.Name()
	config["clean_staging_directory"] = true

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !p.config.CleanStagingDir {
		t.Fatalf("expected clean_staging_directory to be set")
	}
}

func TestProvisionerProvisionDocker_PlaybookFiles(t *testing.T) {
	testProvisionerProvisionDockerWithPlaybookFiles(t, playbookFilesDockerTemplate)
}

func TestProvisionerProvisionDocker_PlaybookFilesWithPlaybookDir(t *testing.T) {
	testProvisionerProvisionDockerWithPlaybookFiles(t, playbookFilesWithPlaybookDirDockerTemplate)
}

func testProvisionerProvisionDockerWithPlaybookFiles(t *testing.T, templateString string) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}

	// this should be a precheck
	cmd := exec.Command("docker", "-v")
	err := cmd.Run()
	if err != nil {
		t.Error("docker command not found; please make sure docker is installed")
	}

	builderT.Test(t, builderT.TestCase{
		Builder:  &docker.Builder{},
		Template: templateString,
		Check: func(a []packersdk.Artifact) error {

			actualContent, err := ioutil.ReadFile("hello_world")
			if err != nil {
				return fmt.Errorf("Expected file not found: %s", err)
			}

			expectedContent := "Hello world!"
			if string(actualContent) != expectedContent {
				return fmt.Errorf(`Unexpected file content: expected="%s", actual="%s"`, expectedContent, actualContent)
			}
			return nil
		},
		Teardown: func() error {
			os.Remove("hello_world")
			return nil
		},
		ProvisionerStore: packersdk.MapOfProvisioner{
			"ansible-local": func() (packersdk.Provisioner, error) { return &Provisioner{}, nil },
			"file":          func() (packersdk.Provisioner, error) { return &file.Provisioner{}, nil },
		},
	})
}

func assertPlaybooksExecuted(comm *communicatorMock, playbooks []string) {
	cmdIndex := 0
	for _, playbook := range playbooks {
		playbook = filepath.ToSlash(playbook)
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
		playbook = filepath.ToSlash(playbook)
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

func assertPlaybooksNotUploaded(comm *communicatorMock, playbooks []string) {
	for _, playbook := range playbooks {
		playbook = filepath.ToSlash(playbook)
		for _, destination := range comm.uploadDestination {
			if strings.HasSuffix(destination, playbook) {
				panic(fmt.Sprintf("Playbook %s was uploaded", playbook))
			}
		}
	}
}

func testConfig() map[string]interface{} {
	m := make(map[string]interface{})
	return m
}

func createTempFile(dir string) string {
	file, err := ioutil.TempFile(dir, "")
	if err != nil {
		panic(fmt.Sprintf("err: %s", err))
	}
	return file.Name()
}

func createTempFiles(dir string, numFiles int) []string {
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
		files = append(files, createTempFile(dir))
	}
	return files
}

func removeFiles(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
}

const playbookFilesDockerTemplate = `
{
	"builders": [
		{
			"type": "test",
			"image": "williamyeh/ansible:centos7",
			"discard": true
		}
	],
	"provisioners": [
		{
			"type": "ansible-local",
			"playbook_files": [
				"test-fixtures/hello.yml",
				"test-fixtures/world.yml"
			]
		},
		{
			"type": "file",
			"source": "/tmp/hello_world",
			"destination": "hello_world",
			"direction": "download"
		}
	]
}
`

const playbookFilesWithPlaybookDirDockerTemplate = `
{
	"builders": [
		{
			"type": "test",
			"image": "williamyeh/ansible:centos7",
			"discard": true
		}
	],
	"provisioners": [
		{
			"type": "ansible-local",
			"playbook_files": [
				"test-fixtures/hello.yml",
				"test-fixtures/world.yml"
			],
			"playbook_dir": "test-fixtures"
		},
		{
			"type": "file",
			"source": "/tmp/hello_world",
			"destination": "hello_world",
			"direction": "download"
		}
	]
}
`
