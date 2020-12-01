package shell_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/helper/tests/acc"
	"github.com/hashicorp/packer/provisioner/shell"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/hashicorp/packer/command"
)

func TestShellLocalProvisionerWithRetryOption(t *testing.T) {
	acc.TestProvisionersPreCheck("shell-local", t)
	acc.TestProvisionersAgainstBuilders(new(ShellLocalProvisionerAccTest), t)
}

type ShellLocalProvisionerAccTest struct{}

func (s *ShellLocalProvisionerAccTest) GetName() string {
	return "file"
}

func (s *ShellLocalProvisionerAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("./test-fixtures", "shell-local-provisioner.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), err
}

func (s *ShellLocalProvisionerAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		"shell-local": func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
	}
}

func (s *ShellLocalProvisionerAccTest) IsCompatible(builder string, vmOS string) bool {
	return vmOS == "linux"
}

func (s *ShellLocalProvisionerAccTest) RunTest(c *command.BuildCommand, args []string) error {
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
