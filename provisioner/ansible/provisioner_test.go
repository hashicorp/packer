// +build !windows

package ansible

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	confighelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
)

// Be sure to remove the Ansible stub file in each test with:
//   defer os.Remove(config["command"].(string))
func testConfig(t *testing.T) map[string]interface{} {
	m := make(map[string]interface{})
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	ansible_stub := path.Join(wd, "packer-ansible-stub.sh")

	err = ioutil.WriteFile(ansible_stub, []byte("#!/usr/bin/env bash\necho ansible 1.6.0"), 0777)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	m["command"] = ansible_stub

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
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should have error")
	}

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["playbook_file"] = playbook_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	err = os.Unsetenv("USER")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_PlaybookFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()

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

func TestProvisionerPrepare_HostKeyFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	filename := make([]byte, 10)
	n, err := io.ReadFull(rand.Reader, filename)
	if n != len(filename) || err != nil {
		t.Fatal("could not create random file name")
	}

	config["ssh_host_key_file"] = fmt.Sprintf("%x", filename)
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["playbook_file"] = playbook_file.Name()

	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should error if ssh_host_key_file does not exist")
	}

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_AuthorizedKeyFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	filename := make([]byte, 10)
	n, err := io.ReadFull(rand.Reader, filename)
	if n != len(filename) || err != nil {
		t.Fatal("could not create random file name")
	}

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["playbook_file"] = playbook_file.Name()
	config["ssh_authorized_key_file"] = fmt.Sprintf("%x", filename)

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should error if ssh_authorized_key_file does not exist")
	}

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	config["ssh_authorized_key_file"] = publickey_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}
}

func TestProvisionerPrepare_LocalPort(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["playbook_file"] = playbook_file.Name()

	config["local_port"] = 65537
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["local_port"] = 22222
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_InventoryDirectory(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := ioutil.TempFile("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := ioutil.TempFile("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := ioutil.TempFile("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["playbook_file"] = playbook_file.Name()

	config["inventory_directory"] = "doesnotexist"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should error if inventory_directory does not exist")
	}

	inventoryDirectory, err := ioutil.TempDir("", "some_inventory_dir")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(inventoryDirectory)

	config["inventory_directory"] = inventoryDirectory
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAnsibleGetVersion(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 and it requires Ansible to be installed")
	}

	var p Provisioner
	p.config.Command = "ansible-playbook"
	err := p.getVersion()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAnsibleGetVersionError(t *testing.T) {
	var p Provisioner
	p.config.Command = "./test-fixtures/exit1"
	err := p.getVersion()
	if err == nil {
		t.Fatal("Should return error")
	}
	if !strings.Contains(err.Error(), "./test-fixtures/exit1 --version") {
		t.Fatal("Error message should include command name")
	}
}

func TestAnsibleLongMessages(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 and it requires Ansible to be installed")
	}

	var p Provisioner
	p.config.Command = "ansible-playbook"
	p.config.PlaybookFile = "./test-fixtures/long-debug-message.yml"
	err := p.Prepare()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &packer.MockCommunicator{}
	ui := &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}

	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestCreateInventoryFile_vers1(t *testing.T) {
	var p Provisioner
	p.Prepare(testConfig(t))
	defer os.Remove(p.config.Command)
	p.ansibleMajVersion = 1
	p.config.User = "testuser"
	p.config.UseProxy = confighelper.TriFalse
	p.generatedData = map[string]interface{}{
		"Host": "123.45.67.89",
		"Port": "2222",
	}

	err := p.createInventoryFile()
	if err != nil {
		t.Fatalf("error creating config using localhost and local port proxy")
	}
	if p.config.InventoryFile == "" {
		t.Fatalf("No inventory file was created")
	}
	defer os.Remove(p.config.InventoryFile)
	f, err := ioutil.ReadFile(p.config.InventoryFile)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}

	expected := "default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=2222\n"
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("File didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
	}
}

func TestCreateInventoryFile_vers2(t *testing.T) {
	var p Provisioner
	p.Prepare(testConfig(t))
	defer os.Remove(p.config.Command)
	p.ansibleMajVersion = 2
	p.config.User = "testuser"
	p.config.UseProxy = confighelper.TriFalse
	p.generatedData = map[string]interface{}{
		"Host": "123.45.67.89",
		"Port": "1234",
	}

	err := p.createInventoryFile()
	if err != nil {
		t.Fatalf("error creating config using localhost and local port proxy")
	}
	if p.config.InventoryFile == "" {
		t.Fatalf("No inventory file was created")
	}
	defer os.Remove(p.config.InventoryFile)
	f, err := ioutil.ReadFile(p.config.InventoryFile)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}
	expected := "default ansible_host=123.45.67.89 ansible_user=testuser ansible_port=1234\n"
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("File didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
	}
}

func TestCreateInventoryFile_Groups(t *testing.T) {
	var p Provisioner
	p.Prepare(testConfig(t))
	defer os.Remove(p.config.Command)
	p.ansibleMajVersion = 1
	p.config.User = "testuser"
	p.config.Groups = []string{"Group1", "Group2"}
	p.config.UseProxy = confighelper.TriFalse
	p.generatedData = map[string]interface{}{
		"Host": "123.45.67.89",
		"Port": "1234",
	}

	err := p.createInventoryFile()
	if err != nil {
		t.Fatalf("error creating config using localhost and local port proxy")
	}
	if p.config.InventoryFile == "" {
		t.Fatalf("No inventory file was created")
	}
	defer os.Remove(p.config.InventoryFile)
	f, err := ioutil.ReadFile(p.config.InventoryFile)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}
	expected := `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group2]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
`
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("File didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
	}
}

func TestCreateInventoryFile_EmptyGroups(t *testing.T) {
	var p Provisioner
	p.Prepare(testConfig(t))
	defer os.Remove(p.config.Command)
	p.ansibleMajVersion = 1
	p.config.User = "testuser"
	p.config.EmptyGroups = []string{"Group1", "Group2"}
	p.config.UseProxy = confighelper.TriFalse
	p.generatedData = map[string]interface{}{
		"Host": "123.45.67.89",
		"Port": "1234",
	}

	err := p.createInventoryFile()
	if err != nil {
		t.Fatalf("error creating config using localhost and local port proxy")
	}
	if p.config.InventoryFile == "" {
		t.Fatalf("No inventory file was created")
	}
	defer os.Remove(p.config.InventoryFile)
	f, err := ioutil.ReadFile(p.config.InventoryFile)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}
	expected := `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
[Group2]
`
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("File didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
	}
}

func TestCreateInventoryFile_GroupsAndEmptyGroups(t *testing.T) {
	var p Provisioner
	p.Prepare(testConfig(t))
	defer os.Remove(p.config.Command)
	p.ansibleMajVersion = 1
	p.config.User = "testuser"
	p.config.Groups = []string{"Group1", "Group2"}
	p.config.EmptyGroups = []string{"Group3"}
	p.config.UseProxy = confighelper.TriFalse
	p.generatedData = map[string]interface{}{
		"Host": "123.45.67.89",
		"Port": "1234",
	}

	err := p.createInventoryFile()
	if err != nil {
		t.Fatalf("error creating config using localhost and local port proxy")
	}
	if p.config.InventoryFile == "" {
		t.Fatalf("No inventory file was created")
	}
	defer os.Remove(p.config.InventoryFile)
	f, err := ioutil.ReadFile(p.config.InventoryFile)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}
	expected := `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group2]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group3]
`
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("File didn't match expected:\n\n file is \n\n %s", f)
	}
}

func TestUseProxy(t *testing.T) {
	type testcase struct {
		UseProxy                   confighelper.Trilean
		generatedData              map[string]interface{}
		expectedSetupAdapterCalled bool
		explanation                string
	}

	tcs := []testcase{
		{
			explanation: "use_proxy is true; we should set up adapter",
			UseProxy:    confighelper.TriTrue,
			generatedData: map[string]interface{}{
				"Host":              "123.45.67.8",
				"Port":              int64(1234),
				"ConnType":          "ssh",
				"SSHPrivateKeyFile": "",
				"SSHPrivateKey":     "asdf",
				"User":              "PartyPacker",
			},
			expectedSetupAdapterCalled: true,
		},
		{
			explanation: "use_proxy is false but no IP addr is available; we should set up adapter anyway.",
			UseProxy:    confighelper.TriFalse,
			generatedData: map[string]interface{}{
				"Host":              "",
				"Port":              nil,
				"ConnType":          "ssh",
				"SSHPrivateKeyFile": "",
				"SSHPrivateKey":     "asdf",
				"User":              "PartyPacker",
			},
			expectedSetupAdapterCalled: true,
		},
		{
			explanation: "use_proxy is false; we shouldn't set up adapter.",
			UseProxy:    confighelper.TriFalse,
			generatedData: map[string]interface{}{
				"Host":              "123.45.67.8",
				"Port":              int64(1234),
				"ConnType":          "ssh",
				"SSHPrivateKeyFile": "",
				"SSHPrivateKey":     "asdf",
				"User":              "PartyPacker",
			},
			expectedSetupAdapterCalled: false,
		},
		{
			explanation: "use_proxy is false but connType isn't ssh or winrm.",
			UseProxy:    confighelper.TriFalse,
			generatedData: map[string]interface{}{
				"Host":              "123.45.67.8",
				"Port":              int64(1234),
				"ConnType":          "docker",
				"SSHPrivateKeyFile": "",
				"SSHPrivateKey":     "asdf",
				"User":              "PartyPacker",
			},
			expectedSetupAdapterCalled: true,
		},
		{
			explanation: "use_proxy is unset; we should default to setting up the adapter (for now).",
			UseProxy:    confighelper.TriUnset,
			generatedData: map[string]interface{}{
				"Host":              "123.45.67.8",
				"Port":              int64(1234),
				"ConnType":          "ssh",
				"SSHPrivateKeyFile": "",
				"SSHPrivateKey":     "asdf",
				"User":              "PartyPacker",
			},
			expectedSetupAdapterCalled: true,
		},
	}

	for _, tc := range tcs {
		var p Provisioner
		p.Prepare(testConfig(t))
		p.config.UseProxy = tc.UseProxy
		defer os.Remove(p.config.Command)
		p.ansibleMajVersion = 1

		var l provisionLogicTracker
		l.setupAdapterCalled = false
		p.setupAdapterFunc = l.setupAdapter
		p.executeAnsibleFunc = l.executeAnsible
		ctx := context.TODO()
		comm := new(packer.MockCommunicator)
		ui := &packer.BasicUi{
			Reader: new(bytes.Buffer),
			Writer: new(bytes.Buffer),
		}
		p.Provision(ctx, ui, comm, tc.generatedData)

		if l.setupAdapterCalled != tc.expectedSetupAdapterCalled {
			t.Fatalf("Should have called set up adapter: %s", tc.explanation)
		}
		os.Remove(p.config.Command)
	}
}
