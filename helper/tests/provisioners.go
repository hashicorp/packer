package testshelper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
)

func getBuilderTestConfig(name string) (string, error) {
	pathName := strings.ReplaceAll(name, "-", "/")
	fileName := name + ".json"
	filePath := filepath.Join("../../builder", pathName, "test-fixtures", fileName)
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), nil
}

func TestProvisionersAgainstBuilders(t *testing.T) {
	// validate if we want to run provisioners acc tests
	b := os.Getenv("ACC_TEST_BUILDERS")
	p := os.Getenv("ACC_TEST_PROVISIONERS")
	if b == "" || p == "" {
		t.Skip("Provisioners Acceptance tests skipped unless env 'ACC_TEST_BUILDERS' and 'ACC_TEST_PROVISIONERS' are set")
	}

	// Get builders and provisioners type to be tested
	builders := strings.Split(b, ",")
	var provisioners []string
	if p == "all" {
		// test all available provisioners
		for k := range ProvisionersAccTest {
			provisioners = append(provisioners, k)
		}
	} else {
		provisioners = strings.Split(p, ",")
	}

	// build template file and run build
	for _, builder := range builders {
		builderConfig, err := getBuilderTestConfig(builder)
		if err != nil {
			t.Fatalf("bad: failed to read builder config: %s", err.Error())
		}
		// Run a build for each builder with each of the provided provisioners
		for _, provisioner := range provisioners {
			testName := fmt.Sprintf("testing %s builder against %s provisioner", builder, provisioner)
			t.Run(testName, func(t *testing.T) {
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
				writeJsonTemplate(out, filePath, t)

				// set pre-config with necessary builder and provisioner
				c := testBuildCommand(t, builder, provisioner)
				args := []string{
					filePath,
				}

				err = provicionerAcc.RunTest(c, args)
				if err != nil {
					// Cleanup created resources
					CleanupFiles(fileName)
					t.Fatalf("bad: failed to to run build: %s", err.Error())
				}

				// Cleanup created resources
				CleanupFiles(fileName)
			})
		}
	}
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

func testBuildCommand(t *testing.T, builder string, provisioner string) *command.BuildCommand {
	c := &command.BuildCommand{
		Meta: TestMetaFile(t),
	}

	c.CoreConfig.Components.BuilderStore = packer.MapOfBuilder{
		builder: func() (packer.Builder, error) { return command.Builders[builder], nil },
	}

	c.CoreConfig.Components.ProvisionerStore = packer.MapOfProvisioner{
		provisioner: func() (packer.Provisioner, error) { return command.Provisioners[provisioner], nil },

		// Add basic provisioner used for testing others provisioners
		"file": func() (packer.Provisioner, error) { return command.Provisioners["file"], nil },
	}

	return c
}

// List of all provisioners available for acceptance test
var ProvisionersAccTest = map[string]ProvisionerAcceptance{}

// List of all builders available for acceptance test
var BuildersAccTest = map[string]BuilderAcceptance{}

type ProvisionerAcceptance interface {
	GetConfig() (string, error)
	RunTest(c *command.BuildCommand, args []string) error
}

type BuilderAcceptance interface {
	GetConfig() (string, error)
	CleanUp() error
}
