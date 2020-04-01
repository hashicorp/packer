package acceptance_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	amazonEBS "github.com/hashicorp/packer/builder/amazon/ebs/acceptance"
	virtualboxISO "github.com/hashicorp/packer/builder/virtualbox/iso/acceptance"
	shell "github.com/hashicorp/packer/provisioner/shell/acceptance"

	"github.com/hashicorp/packer/command"
	testshelper "github.com/hashicorp/packer/helper/tests"
	"github.com/hashicorp/packer/packer"
)

func TestProvisionersAgainstBuilders(t *testing.T) {
	b, p := provisionerAccTestPreCheck(t)

	// Get builders and provisioners type to be tested
	var builders []string
	for k := range BuildersAccTest {
		// This will validate that only defined builders are executed against
		if b != "all" && !strings.Contains(b, k) {
			continue
		}
		builders = append(builders, k)
	}

	var provisioners []string
	for k := range ProvisionersAccTest {
		if p != "all" && !strings.Contains(p, k) {
			continue
		}
		provisioners = append(provisioners, k)
	}

	// build template file and run build
	for _, builder := range builders {
		builderAcc := BuildersAccTest[builder]
		builderConfig, err := builderAcc.GetConfig()
		if err != nil {
			t.Fatalf("bad: failed to read builder config: %s", err.Error())
		}
		// Run a build for each builder with each of the provided provisioners
		for _, provisioner := range provisioners {
			testName := fmt.Sprintf("testing %s builder against %s provisioner", builder, provisioner)
			t.Run(testName, func(t *testing.T) {
				provisionerAcc := ProvisionersAccTest[provisioner]
				provisionerConfig, err := provisionerAcc.GetConfig()
				if err != nil {
					t.Fatalf("bad: failed to read provisioner config: %s", err.Error())
				}

				// Write json template
				out := bytes.NewBuffer(nil)
				fmt.Fprintf(out, `{"builders": [%s],"provisioners": [%s]}`, builderConfig, provisionerConfig)
				fileName := fmt.Sprintf("%s_%s.json", builder, provisioner)
				filePath := filepath.Join("./", fileName)
				writeJsonTemplate(out, filePath, t)

				// set pre-config with necessary builder and provisioner
				c := buildCommand(t, builderAcc, provisionerAcc)
				args := []string{
					filePath,
				}

				err = provisionerAcc.RunTest(c, args)
				if err != nil {
					// Cleanup created resources
					testshelper.CleanupFiles(fileName)
					builderAcc.CleanUp()
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

func provisionerAccTestPreCheck(t *testing.T) (string, string) {
	b := os.Getenv("ACC_TEST_BUILDERS")
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	// validate if we want to run provisioners acc tests
	if b == "" || p == "" {
		t.Skip("Provisioners Acceptance tests skipped unless env 'ACC_TEST_BUILDERS' and 'ACC_TEST_PROVISIONERS' are set")
	}
	return b, p
}

func writeJsonTemplate(out *bytes.Buffer, filePath string, t *testing.T) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("bad: failed to create template file: %s", err.Error())
	}
	_, err = outputFile.Write(out.Bytes())
	if err != nil {
		t.Fatalf("bad: failed to write template file: %s", err.Error())
	}
	outputFile.Sync()
}

func buildCommand(t *testing.T, builder BuilderAcceptance, provisioner ProvisionerAcceptance) *command.BuildCommand {
	c := &command.BuildCommand{
		Meta: testshelper.TestMetaFile(t),
	}
	c.CoreConfig.Components.BuilderStore = builder.GetBuilderStore()
	c.CoreConfig.Components.ProvisionerStore = provisioner.GetProvisionerStore()

	return c
}

// List of all provisioners available for acceptance test
var ProvisionersAccTest = map[string]ProvisionerAcceptance{
	"shell": new(shell.ShellProvisionerAccTest),
}

// List of all builders available for acceptance test
var BuildersAccTest = map[string]BuilderAcceptance{
	"virtualbox-iso": new(virtualboxISO.VirtualBoxISOAccTest),
	"amazon-ebs":     new(amazonEBS.AmazonEBSAccTest),
}

type ProvisionerAcceptance interface {
	GetConfig() (string, error)
	RunTest(c *command.BuildCommand, args []string) error
	GetProvisionerStore() packer.MapOfProvisioner
}

type BuilderAcceptance interface {
	GetConfig() (string, error)
	CleanUp() error
	GetBuilderStore() packer.MapOfBuilder
}
