package iso

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
)

func TestBuilderAcc_basic(t *testing.T) {
	templatePath := filepath.Join("testdata", "minimal.json")
	bytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to load template file %s", templatePath)
	}

	testCase := &acctest.PluginTestCase{
		Name: "vmware-iso_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			testutils.CleanupFiles("output-vmware-iso", "packer_cache")
			return nil
		},
		Template: string(bytes),
		Type:     "vmware-iso",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
