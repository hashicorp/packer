package acceptance

import (
	"fmt"
	"os"
	"testing"
)

const TestEnvVar = "PACKER_ACC"

func AccTestPreValidate(t *testing.T) {
	// We only run acceptance tests if an env var is set because they're
	// slow and generally require some outside configuration.
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
		return
	}

	// We require verbose mode so that the user knows what is going on.
	if !testing.Verbose() {
		t.Fatal("Acceptance tests must be run with the -v flag on tests")
		return
	}
}
