package acctest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestEnvVar must be set to a non-empty value for acceptance tests to run.
const TestEnvVar = "PACKER_ACC"

// PluginTestCase is a single set of tests to run for a plugin.
// A PluginTestCase should generally map 1:1 to each test method for your
// acceptance tests.
// Requirements:
// - If not using 'packer init', the plugin must be previously installed
// - Packer must be installed locally
type PluginTestCase struct {
	// Init, if true `packer init` will be executed prior to `packer build`.
	Init bool
	// CheckInit is called after packer init step is executed in order to test that
	// the step executed successfully. If this is not set, then the next
	// step will be called
	CheckInit func(*exec.Cmd, string) error

	// Check is called after this step is executed in order to test that
	// the step executed successfully. If this is not set, then the next
	// step will be called
	Check func(*exec.Cmd, string) error
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
	Teardown TestTeardownFunc
	// Template is the testing HCL2 template to use.
	Template string
	// Type is the type of the plugin.
	Type string
}

// TestTeardownFunc is the callback used for Teardown in TestCase.
type TestTeardownFunc func() error

//nolint:errcheck
func TestPlugin(t *testing.T, testCase *PluginTestCase) {
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
		return
	}

	if testCase.Setup != nil {
		err := testCase.Setup()
		if err != nil {
			t.Fatalf("test %s setup failed: %s", testCase.Name, err)
		}
	}

	logfile := fmt.Sprintf("packer_log_%s.txt", testCase.Name)

	extension := ".pkr.hcl"
	if err := json.Unmarshal([]byte(testCase.Template), &(map[string]interface{}{})); err == nil {
		extension = ".json"
	}
	templatePath := fmt.Sprintf("./%s%s", testCase.Name, extension)

	// Write config hcl2 template
	out := bytes.NewBuffer(nil)
	fmt.Fprintf(out, testCase.Template)
	outputFile, err := os.Create(templatePath)
	if err != nil {
		t.Fatalf("bad: failed to create template file: %s", err.Error())
	}
	_, err = outputFile.Write(out.Bytes())
	if err != nil {
		t.Fatalf("bad: failed to write template file: %s", err.Error())
	}
	outputFile.Sync()

	// Make sure packer is installed:
	packerbin, err := exec.LookPath("packer")
	if err != nil {
		t.Fatalf("Couldn't find packer binary installed on system: %s", err.Error())
	}

	if testCase.Init {
		initLogfile := fmt.Sprintf("packer_init_log_%s.txt", testCase.Name)
		initCommand := exec.Command(packerbin, "init", templatePath)
		initCommand.Env = append(initCommand.Env, os.Environ()...)
		initCommand.Env = append(initCommand.Env, "PACKER_LOG=1", fmt.Sprintf("PACKER_LOG_PATH=%s", initLogfile))
		initCommand.Run()

		if testCase.CheckInit != nil {
			if err := testCase.CheckInit(initCommand, initLogfile); err != nil {
				cwd, _ := os.Getwd()
				t.Fatalf(fmt.Sprintf("Error running plugin acceptance"+
					" tests: %s\nLogs can be found at %s\nand the "+
					"acceptance test template can be found at %s",
					err.Error(), filepath.Join(cwd, initLogfile),
					filepath.Join(cwd, templatePath)))
			} else {
				os.Remove(initLogfile)
			}
		}
	}

	// Run build
	buildCommand := exec.Command(packerbin, "build", "--machine-readable", templatePath)
	buildCommand.Env = append(buildCommand.Env, os.Environ()...)
	buildCommand.Env = append(buildCommand.Env, "PACKER_LOG=1",
		fmt.Sprintf("PACKER_LOG_PATH=%s", logfile))
	buildCommand.Run()

	// Check for test custom pass/fail before we clean up
	var checkErr error
	if testCase.Check != nil {
		checkErr = testCase.Check(buildCommand, logfile)
	}
	// Clean up anything created in the plugin run
	if testCase.Teardown != nil {
		cleanErr := testCase.Teardown()
		if cleanErr != nil {
			t.Logf("bad: failed to clean up test-created resources: %s", cleanErr.Error())
		}
	}

	// Fail test if check failed.
	if checkErr != nil {
		cwd, _ := os.Getwd()
		t.Fatalf(fmt.Sprintf("Error running plugin acceptance"+
			" tests: %s\nLogs can be found at %s\nand the "+
			"acceptance test template can be found at %s",
			checkErr.Error(), filepath.Join(cwd, logfile),
			filepath.Join(cwd, templatePath)))
	} else {
		os.Remove(templatePath)
		os.Remove(logfile)
	}
}
