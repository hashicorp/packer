package shell_acceptance

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	testshelper "github.com/hashicorp/packer/helper/tests"
)

type ShellProvisionerAccTest struct{}

func (s *ShellProvisionerAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("../../../provisioner/shell/acceptance/test-fixtures/", "shell-provisioner.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), nil
}

func (s *ShellProvisionerAccTest) RunTest(c *command.BuildCommand, args []string) error {
	UUID, _ := uuid.GenerateUUID()
	os.Setenv("PACKER_RUN_UUID", UUID)

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
