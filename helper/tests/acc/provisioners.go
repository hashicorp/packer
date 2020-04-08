package acc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testshelper "github.com/hashicorp/packer/helper/tests"

	amazonEBS "github.com/hashicorp/packer/builder/amazon/ebs/acceptance"
	virtualboxISO "github.com/hashicorp/packer/builder/virtualbox/iso/acceptance"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
)

func TestProvisionersAgainstBuilders(provisionerAcc ProvisionerAcceptance, t *testing.T) {
	provisioner := provisionerAcc.GetName()
	builders := checkBuilders(t)

	// build template file and run a build for each builder with the provisioner
	for _, builder := range builders {
		builderAcc := BuildersAccTest[builder]
		builderConfigs, err := builderAcc.GetConfigs()
		if err != nil {
			t.Fatalf("bad: failed to read builder config: %s", err.Error())
		}

		for vmOS, builderConfig := range builderConfigs {
			if !provisionerAcc.IsCompatible(builder, vmOS) {
				continue
			}

			testName := fmt.Sprintf("testing %s builder against %s provisioner", builder, provisioner)
			t.Run(testName, func(t *testing.T) {
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
				// Cleanup created resources
				testshelper.CleanupFiles(fileName)
				cleanErr := builderAcc.CleanUp()
				if cleanErr != nil {
					log.Printf("bad: failed to clean up resources: %s", cleanErr.Error())
				}
				if err != nil {
					t.Fatalf("bad: failed to to run build: %s", err.Error())
				}
			})
		}
	}
}

// TestProvisionersPreCheck checks if the Provisioner with name is set in ACC_TEST_PROVISIONERS environment variable
func TestProvisionersPreCheck(name string, t *testing.T) {
	p := os.Getenv("ACC_TEST_PROVISIONERS")

	if p == "all" {
		return
	}

	provisioners := strings.Split(p, ",")
	for _, provisioner := range provisioners {
		if provisioner == name {
			return
		}
	}

	msg := fmt.Sprintf("Provisioner %q not defined in ACC_TEST_PROVISIONERS", name)
	t.Skip(msg)

}

func checkBuilders(t *testing.T) []string {
	b := os.Getenv("ACC_TEST_BUILDERS")
	// validate if we want to run provisioners acc tests
	if b == "" {
		t.Skip("Provisioners Acceptance tests skipped unless env 'ACC_TEST_BUILDERS' is set")
	}

	// Get builders type to test provisioners against
	var builders []string
	for k := range BuildersAccTest {
		// This will validate that only defined builders are executed against
		if b != "all" && !strings.Contains(b, k) {
			continue
		}
		builders = append(builders, k)
	}
	return builders
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

type ProvisionerAcceptance interface {
	GetName() string
	GetConfig() (string, error)
	GetProvisionerStore() packer.MapOfProvisioner
	IsCompatible(builder string, vmOS string) bool
	RunTest(c *command.BuildCommand, args []string) error
}

type BuilderAcceptance interface {
	GetConfigs() (map[string]string, error)
	GetBuilderStore() packer.MapOfBuilder
	CleanUp() error
}

// List of all builders available for acceptance test
var BuildersAccTest = map[string]BuilderAcceptance{
	"virtualbox-iso": new(virtualboxISO.VirtualBoxISOAccTest),
	"amazon-ebs":     new(amazonEBS.AmazonEBSAccTest),
}
