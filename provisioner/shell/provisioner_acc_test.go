package shell_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/helper/tests/acc"
	"github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	testshelper "github.com/hashicorp/packer/helper/tests"
)

func TestShellProvisioner(t *testing.T) {
	acc.TestProvisionersPreCheck("shell", t)
	acc.TestProvisionersAgainstBuilders(new(ShellProvisionerAccTest), t)
}

type ShellProvisionerAccTest struct{}

func (s *ShellProvisionerAccTest) GetName() string {
	return "shell"
}

func (s *ShellProvisionerAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("./test-fixtures", "shell-provisioner.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), err
}

func (s *ShellProvisionerAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		"shell": func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
		"file":  func() (packer.Provisioner, error) { return &file.Provisioner{}, nil },
	}
}

func (s *ShellProvisionerAccTest) IsCompatible(builder string, vmOS string) bool {
	return vmOS == "linux"
}

func (s *ShellProvisionerAccTest) RunTest(c *command.BuildCommand, args []string) error {
	UUID := os.Getenv("PACKER_RUN_UUID")
	if UUID == "" {
		UUID, _ = uuid.GenerateUUID()
		os.Setenv("PACKER_RUN_UUID", UUID)
	}

	file := "provisioner.shell." + UUID + ".txt"
	defer testshelper.CleanupFiles(file)

	if code := c.Run(args); code != 0 {
		ui := c.Meta.Ui.(*packersdk.BasicUi)
		out := ui.Writer.(*bytes.Buffer)
		err := ui.ErrorWriter.(*bytes.Buffer)
		return fmt.Errorf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			out.String(),
			err.String())
	}

	if !testshelper.FileExists(file) {
		return fmt.Errorf("Expected to find %s", file)
	}
	return nil
}
