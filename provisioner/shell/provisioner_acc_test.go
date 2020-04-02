package shell_test

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/packer/helper/tests/acc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	testshelper "github.com/hashicorp/packer/helper/tests"
)

func TestShellProvisioner(t *testing.T) {
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	if p != "all" && !strings.Contains(p, "shell") {
		t.Skip()
	}
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
	return string(file), nil
}

func (s *ShellProvisionerAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		"shell": func() (packer.Provisioner, error) { return command.Provisioners["shell"], nil },
		"file":  func() (packer.Provisioner, error) { return command.Provisioners["file"], nil },
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
		ui := c.Meta.Ui.(*packer.BasicUi)
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
