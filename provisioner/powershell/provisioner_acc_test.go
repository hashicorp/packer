package powershell_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/helper/tests/acc"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/provisioner/powershell"
	windowsshellprovisioner "github.com/hashicorp/packer/provisioner/windows-shell"
)

const TestProvisionerName = "powershell"

func TestAccPowershellProvisioner_basic(t *testing.T) {
	acc.TestProvisionersPreCheck(TestProvisionerName, t)

	testProvisioner := PowershellProvisionerAccTest{"powershell-provisioner-cleanup.txt"}
	acc.TestProvisionersAgainstBuilders(&testProvisioner, t)
}

func TestAccPowershellProvisioner_Inline(t *testing.T) {
	acc.TestProvisionersPreCheck(TestProvisionerName, t)

	testProvisioner := PowershellProvisionerAccTest{"powershell-inline-provisioner.txt"}
	acc.TestProvisionersAgainstBuilders(&testProvisioner, t)
}

func TestAccPowershellProvisioner_Script(t *testing.T) {
	acc.TestProvisionersPreCheck(TestProvisionerName, t)

	testProvisioner := PowershellProvisionerAccTest{"powershell-script-provisioner.txt"}
	acc.TestProvisionersAgainstBuilders(&testProvisioner, t)
}

type PowershellProvisionerAccTest struct {
	ConfigName string
}

func (s *PowershellProvisionerAccTest) GetName() string {
	return TestProvisionerName
}

func (s *PowershellProvisionerAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("./test-fixtures", s.ConfigName)
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("os.Open:%v", err)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	if err != nil {
		return "", fmt.Errorf("ioutil.ReadAll:%v", err)
	}
	return string(file), nil
}

func (s *PowershellProvisionerAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		TestProvisionerName: func() (packer.Provisioner, error) { return &powershell.Provisioner{}, nil },
		"windows-shell":     func() (packer.Provisioner, error) { return &windowsshellprovisioner.Provisioner{}, nil },
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
		ui := c.Meta.Ui.(*packersdk.BasicUi)
		out := ui.Writer.(*bytes.Buffer)
		err := ui.ErrorWriter.(*bytes.Buffer)
		return fmt.Errorf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			out.String(),
			err.String())
	}

	return nil
}
