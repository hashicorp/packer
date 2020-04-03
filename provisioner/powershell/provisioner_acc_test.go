package powershell_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/helper/tests/acc"
	"github.com/hashicorp/packer/packer"
)

func TestPowershellProvisioner_Inline(t *testing.T) {
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	if p != "all" && !strings.Contains(p, "powershell") {
		t.Skip()
	}

	testProvisioner := PowershellProvisionerAccTest{"powershell-inline-provisioner.txt"}
	acc.TestProvisionersAgainstBuilders(&testProvisioner, t)
}

func TestPowershellProvisioner_Script(t *testing.T) {
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	if p != "all" && !strings.Contains(p, "powershell") {
		t.Skip()
	}

	testProvisioner := PowershellProvisionerAccTest{"powershell-script-provisioner.txt"}
	acc.TestProvisionersAgainstBuilders(&testProvisioner, t)
}

type PowershellProvisionerAccTest struct {
	ConfigName string
}

func (s *PowershellProvisionerAccTest) GetName() string {
	return "powershell"
}

func (s *PowershellProvisionerAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("./test-fixtures", s.ConfigName)
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), nil
}

func (s *PowershellProvisionerAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		"powershell": func() (packer.Provisioner, error) { return command.Provisioners["powershell"], nil },
	}
}

func (s *PowershellProvisionerAccTest) IsCompatible(builder string, vmOS string) bool {
	return vmOS == "windows"
}

func (s *PowershellProvisionerAccTest) RunTest(c *command.BuildCommand, args []string) error {
	UUID := os.Getenv("PACKER_RUN_UUID")
	if UUID == "" {
		UUID, _ = uuid.GenerateUUID()
		os.Setenv("PACKER_RUN_UUID", UUID)
	}

	if code := c.Run(args); code != 0 {
		ui := c.Meta.Ui.(*packer.BasicUi)
		out := ui.Writer.(*bytes.Buffer)
		err := ui.ErrorWriter.(*bytes.Buffer)
		return fmt.Errorf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			out.String(),
			err.String())
	}

	return nil
}
