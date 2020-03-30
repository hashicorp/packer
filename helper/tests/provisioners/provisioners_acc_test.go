package acceptance

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	amazonEbs "github.com/hashicorp/packer/builder/amazon/ebs/acceptance"
	"github.com/hashicorp/packer/command"
	testshelper "github.com/hashicorp/packer/helper/tests"
	"github.com/hashicorp/packer/packer"
	shell "github.com/hashicorp/packer/provisioner/shell/acceptance"
)

func TestProvisionersAgainstBuilders(t *testing.T) {
	// validate if we want to run provisioners acc tests
	b := os.Getenv("ACC_TEST_BUILDERS")
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	if b == "" || p == "" {
		t.Skip("Provisioners Acceptance tests skipped unless env 'ACC_TEST_BUILDERS' and 'ACC_TEST_PROVISIONERS' are set")
	}

	// set pre-config with necessary builders and provisioners
	c := &command.BuildCommand{
		Meta: testshelper.TestMetaFile(t),
	}

	builders := strings.Split(b, ",")
	mapOfBuilders := packer.MapOfBuilder{}
	for _, builder := range builders {
		mapOfBuilders[builder] = func() (packer.Builder, error) { return command.Builders[builder], nil }
	}

	provisioners := strings.Split(p, ",")
	mapOfProvisioner := packer.MapOfProvisioner{}
	for _, provisioner := range provisioners {
		mapOfProvisioner[provisioner] = func() (packer.Provisioner, error) { return command.Provisioners[provisioner], nil }
	}

	// Add basic provisioner used for testing others provisioners
	mapOfProvisioner["file"] = func() (packer.Provisioner, error) { return command.Provisioners["file"], nil }

	c.CoreConfig.Components.BuilderStore = mapOfBuilders
	c.CoreConfig.Components.ProvisionerStore = mapOfProvisioner

	// build template file and run build
	for _, builder := range builders {
		builderAcc := BuildersAccTest[builder]
		builderConfig, err := builderAcc.GetConfig()
		if err != nil {
			t.Fatalf("bad: failed to read builder config: %s", err.Error())
		}

		// Run a build for each builder with each of the provided provisioners
		for _, provisioner := range provisioners {
			provicionerAcc := ProvisionersAccTest[provisioner]
			provisionerConfig, err := provicionerAcc.GetConfig()
			if err != nil {
				t.Fatalf("bad: failed to read provisioner config: %s", err.Error())
			}

			// Write json template
			out := bytes.NewBuffer(nil)
			fmt.Fprintf(out, `{"builders": [%s],"provisioners": [%s]}`, builderConfig, provisionerConfig)
			fileName := fmt.Sprintf("%s_%s.json", builder, provisioner)
			filePath := filepath.Join("./", fileName)
			outputFile, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("bad: failed to create template file: %s", err.Error())
			}
			_, err = outputFile.Write(out.Bytes())
			if err != nil {
				t.Fatalf("bad: failed to write template file: %s", err.Error())
			}
			outputFile.Sync()

			// Run test
			args := []string{
				filePath,
			}
			testName := fmt.Sprintf("testing %s agaist %s", builder, provisioner)
			t.Run(testName, func(t *testing.T) {
				err = provicionerAcc.RunTest(c, args)
				if err != nil {
					t.Fatalf("bad: failed to to run build: %s", err.Error())
				}

				// Cleanup created resources
				testshelper.CleanupFiles(fileName)
				err = builderAcc.CleanUp()
				if err != nil {
					t.Fatalf("bad: failed to clean up resources: %s", err.Error())
				}
			})
		}
	}
}

// List of all provisioners available for acceptance test
var ProvisionersAccTest = map[string]ProvisionerAcceptance{
	"shell": new(shell.ShellProvisionerAccTest),
}

// List of all builders available for acceptance test
var BuildersAccTest = map[string]BuilderAcceptance{
	"amazon-ebs": new(amazonEbs.AmazonEBSAccTest),
}

type ProvisionerAcceptance interface {
	GetConfig() (string, error)
	RunTest(c *command.BuildCommand, args []string) error
}

type BuilderAcceptance interface {
	GetConfig() (string, error)
	CleanUp() error
}
