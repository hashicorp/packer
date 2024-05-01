package test

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// BuildTestPacker builds a new Packer binary based on the current state of the repository.
//
// If for some reason the binary cannot be built, we will immediately exit with an error.
func BuildTestPacker(t *testing.T) (string, error) {
	testDir, err := currentDir()
	if err != nil {
		return "", fmt.Errorf("failed to compile packer binary: %s", err)
	}

	packerCoreDir := filepath.Dir(testDir)

	outBin := filepath.Join(os.TempDir(), fmt.Sprintf("packer_core-%d", rand.Int()))

	compileCommand := exec.Command("go", "build", "-C", packerCoreDir, "-o", outBin)
	logs, err := compileCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile Packer core: %s\ncompilation logs: %s", err, logs)
	}

	return outBin, nil
}
