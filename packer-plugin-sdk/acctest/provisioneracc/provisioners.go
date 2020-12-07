package provisioneracc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	builderT "github.com/hashicorp/packer/packer-plugin-sdk/acctest"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// ProvisionerTestCase is a single set of tests to run for a provisioner.
// A ProvisionerTestCase should generally map 1:1 to each test method for your
// acceptance tests.
type ProvisionerTestCase struct {
	// Check is called after this step is executed in order to test that
	// the step executed successfully. If this is not set, then the next
	// step will be called
	Check func(*exec.Cmd, string) error
	// IsCompatible checks whether a provisioner is able to run against a
	// given builder type and guest operating system, and returns a boolean.
	// if it returns true, the test combination is okay to run. If false, the
	// test combination is not okay to run.
	IsCompatible func(builderType string, BuilderGuestOS string) bool
	// Name is the name of the test case. Be simple but unique and descriptive.
	Name string
	// Setup, if non-nil, will be called once before the test case
	// runs. This can be used for some setup like setting environment
	// variables, or for validation prior to the
	// test running. For example, you can use this to make sure certain
	// binaries are installed, or text fixtures are in place.
	Setup func() error
	// Teardown will be called before the test case is over regardless
	// of if the test succeeded or failed. This should return an error
	// in the case that the test can't guarantee all resources were
	// properly cleaned up.
	Teardown builderT.TestTeardownFunc
	// Template is the provisioner template to use.
	// The provisioner template fragment must be a json-formatted string
	// containing the provisioner definition but no other portions of a packer
	// template. For
	// example:
	//
	// ```json
	// {
	// 	"type": "shell-local",
	// 	"inline", ["echo hello world"]
	// }
	//```
	//
	// is a valid entry for "template" here, but the complete Packer template:
	//
	// ```json
	// {
	// 	"provisioners": [
	// 		{
	// 			"type": "shell-local",
	// 			"inline", ["echo hello world"]
	// 		}
	// 	]
	// }
	// ```
	//
	// is invalid as input.
	//
	// You may provide multiple provisioners in the same template. For example:
	// ```json
	// {
	// 	"type": "shell-local",
	// 	"inline", ["echo hello world"]
	// },
	// {
	// 	"type": "shell-local",
	// 	"inline", ["echo hello world 2"]
	// }
	// ```
	Template string
	// Type is the type of provisioner.
	Type string
}

// BuilderFixtures are basic builder test configurations and metadata used
// in provisioner acceptance testing. These are frameworks to be used by
// provisioner tests, not tests in and of themselves. BuilderFixtures should
// generally be simple and not contain excessive or complex configurations.
// Instantiations of this struct are stored in the builders.go file in this
// module.
type BuilderFixture struct {
	// Name is the name of the builder fixture.
	// Be simple and descriptive.
	Name string
	// Setup creates necessary extra test fixtures, and renders their values
	// into the BuilderFixture.Template.
	Setup func()
	// Template is the path to a builder template fragment.
	// The builder template fragment must be a json-formatted file containing
	// the builder definition but no other portions of a packer template. For
	// example:
	//
	// ```json
	// {
	// 	"type": "null",
	// 	"communicator", "none"
	// }
	//```
	//
	// is a valid entry for "template" here, but the complete Packer template:
	//
	// ```json
	// {
	// 	"builders": [
	// 		"type": "null",
	// 		"communicator": "none"
	// 	]
	// }
	// ```
	//
	// is invalid as input.
	//
	// Only provide one builder template fragment per file.
	TemplatePath string

	// GuestOS says what guest os type the builder template fragment creates.
	// Valid values are "windows", "linux" or "darwin" guests.
	GuestOS string

	// HostOS says what host os type the builder is capable of running on.
	// Valid values are "any", windows", or "posix". If you set "posix", then
	// this builder can run on a "linux" or "darwin" platform. If you set
	// "any", then this builder can be used on any platform.
	HostOS string

	Teardown builderT.TestTeardownFunc
}

func fixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "test-fixtures")
}

func LoadBuilderFragment(templateFragmentPath string) (string, error) {
	dir := fixtureDir()
	fragmentAbsPath := filepath.Join(dir, templateFragmentPath)
	fragmentFile, err := os.Open(fragmentAbsPath)
	if err != nil {
		return "", fmt.Errorf("Unable find %s", fragmentAbsPath)
	}
	defer fragmentFile.Close()

	fragmentString, err := ioutil.ReadAll(fragmentFile)
	if err != nil {
		return "", fmt.Errorf("Unable to read %s", fragmentAbsPath)
	}

	return string(fragmentString), nil
}

func RunProvisionerAccTest(testCase *ProvisionerTestCase, t *testing.T) {
	TestProvisionersAgainstBuilders(testCase, t)
}

//nolint:errcheck
func TestProvisionersAgainstBuilders(testCase *ProvisionerTestCase, t *testing.T) {
	// retrieve user-desired builders.
	builderTypes := checkBuilders(t)

	// Run this provisioner test case against each builder type requested.
	for _, builderType := range builderTypes {
		buildFixtures := BuildersAccTest[builderType]
		// loop over individual build templates, merge with provisioner
		// templates, and shell out to run test.
		for _, buildFixture := range buildFixtures {
			if !testCase.IsCompatible(builderType, buildFixture.GuestOS) {
				continue
			}

			testName := fmt.Sprintf("%s on %s", testCase.Name, buildFixture.Name)

			if testCase.Setup != nil {
				err := testCase.Setup()
				if err != nil {
					t.Fatalf("test %s setup failed: %s", testName, err)
				}
			}

			t.Run(testName, func(t *testing.T) {
				builderFragment, err := LoadBuilderFragment(buildFixture.TemplatePath)
				if err != nil {
					t.Fatalf("failed to load builder fragment: %s", err)
				}

				// Combine provisioner and builder template fragments; write to
				// file.
				out := bytes.NewBuffer(nil)
				fmt.Fprintf(out, `{"builders": [%s],"provisioners": [%s]}`,
					builderFragment, testCase.Template)
				templateName := fmt.Sprintf("%s_%s.json", builderType, testCase.Type)
				templatePath := filepath.Join("./", templateName)
				writeJsonTemplate(out, templatePath, t)
				logfile := fmt.Sprintf("packer_log_%s_%s.txt", builderType, testCase.Type)

				// Run build
				buildCommand := exec.Command("packer", "build", "--machine-readable", templatePath)
				buildCommand.Env = append(buildCommand.Env, os.Environ()...)
				buildCommand.Env = append(buildCommand.Env, "PACKER_LOG=1",
					fmt.Sprintf("PACKER_LOG_PATH=%s", logfile))
				buildCommand.Run()

				// Check for test custom pass/fail before we clean up
				var checkErr error
				if testCase.Check != nil {
					checkErr = testCase.Check(buildCommand, logfile)
				}

				// Cleanup stuff created by builder.
				cleanErr := buildFixture.Teardown()
				if cleanErr != nil {
					log.Printf("bad: failed to clean up builder-created resources: %s", cleanErr.Error())
				}
				// Clean up anything created in provisioner run
				if testCase.Teardown != nil {
					cleanErr = testCase.Teardown()
					if cleanErr != nil {
						log.Printf("bad: failed to clean up test-created resources: %s", cleanErr.Error())
					}
				}

				// Fail test if check failed.
				if checkErr != nil {
					t.Fatalf(fmt.Sprint("Error running provisioner acceptance"+
						" tests: %s\nLogs can be found at %s and the "+
						"acceptance test template can be found at %s",
						checkErr.Error(), logfile, templatePath))
				} else {
					os.Remove(templatePath)
					os.Remove(logfile)
				}
			})
		}
	}
}

// checkBuilders retrieves  all of the builders that the user has requested to
// run acceptance tests against.
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

// BuilderAcceptance is specialized tooling implemented by individual builders
// To add your builder to the provisioner testing framework, create a struct
// that implements the this interface, add it to the BuildersAccTest map below.
// TODO add this interface to the plugin server so that Packer can request it
// From the plugin rather than importing it here.
type BuilderAcceptance interface {
	// GetConfigs provides a mapping of guest OS architecture to builder
	// template fragment.
	// The builder template fragment must be a json-formatted string containing
	// the builder definition but no other portions of a packer template. For
	// example:
	//
	// ```json
	// {
	// 	"type": "null",
	// 	"communicator", "none"
	// }
	//```
	//
	// is a valid entry for "template" here, but the complete Packer template:
	//
	// ```json
	// {
	// 	"builders": [
	// 		"type": "null",
	// 		"communicator": "none"
	// 	]
	// }
	// ```
	//
	// is invalid as input.
	//
	// Valid keys for the map are "linux" and "windows". These keys will be used
	// to determine whether a given builder template is compatible with a given
	// provisioner template.
	GetConfigs() (map[string]string, error)
	// GetBuilderStore() returns a MapOfBuilder that contains the actual builder
	// struct definition being used for this test.
	GetBuilderStore() packersdk.MapOfBuilder
	// CleanUp cleans up any side-effects of the builder not already cleaned up
	// by the builderT framework.
	CleanUp() error
}

// Mapping of all builder fixtures defined for a given builder type.
var BuildersAccTest = map[string][]*BuilderFixture{
	"virtualbox-iso": []*BuilderFixture{VirtualboxBuilderFixtureWindows},
	"amazon-ebs":     []*BuilderFixture{AmasonEBSBuilderFixtureLinux, AmasonEBSBuilderFixtureWindows},
}
