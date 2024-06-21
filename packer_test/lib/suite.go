package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/suite"
)

type PackerTestSuite struct {
	suite.Suite
	// pluginsDirectory is the directory in which plugins are compiled.
	//
	// Those binaries are not necessarily meant to be used as-is, but
	// instead should be used for composing plugin installation directories.
	pluginsDirectory string
	// packerPath is the location in which the Packer executable is compiled
	//
	// Since we don't necessarily want to manually compile Packer beforehand,
	// we compile it on demand, and use this executable for the tests.
	packerPath string
}

func (ts *PackerTestSuite) buildPluginVersion(waitgroup *sync.WaitGroup, versionString string, t *testing.T) {
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		ts.BuildSimplePlugin(versionString, t)
	}()
}

func (ts *PackerTestSuite) CompileTestPluginVersions(t *testing.T, versions ...string) {
	wg := &sync.WaitGroup{}

	for _, ver := range versions {
		ts.buildPluginVersion(wg, ver, t)
	}

	wg.Wait()
}

// BuildSimplePlugin creates a plugin that essentially does nothing.
//
// The plugin's code is contained in a subdirectory of this, and lets us
// change the attributes of the plugin binary itself, like the SDK version,
// the plugin's version, etc.
//
// The plugin is functional, and can be used to run builds with.
// There won't be anything substantial created though, its goal is only
// to validate the core functionality of Packer.
//
// The path to the plugin is returned, it won't be removed automatically
// though, deletion is the caller's responsibility.
func (ts *PackerTestSuite) BuildSimplePlugin(versionString string, t *testing.T) string {
	// Only build plugin binary if not already done beforehand
	path, ok := LoadPluginVersion(versionString)
	if ok {
		return path
	}

	v := version.Must(version.NewSemver(versionString))

	t.Logf("Building plugin in version %v", v)

	testDir, err := currentDir()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s", err)
	}

	testerPluginDir := filepath.Join(testDir, "plugin_tester")
	outBin := filepath.Join(ts.pluginsDirectory, BinaryName(v))

	compileCommand := exec.Command("go", "build", "-C", testerPluginDir, "-o", outBin, "-ldflags", LDFlags(v), ".")
	logs, err := compileCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s\ncompiler logs: %s", err, logs)
	}

	StorePluginVersion(v.String(), outBin)

	return outBin
}

// SkipNoAcc is a pre-condition that skips the test if the PACKER_ACC environment
// variable is unset, or set to "0".
//
// This allows us to build tests with a potential for long runs (or errors like
// rate-limiting), so we can still test them, but only in a longer timeouted
// context.
func (ts *PackerTestSuite) SkipNoAcc() {
	acc := os.Getenv("PACKER_ACC")
	if acc == "" || acc == "0" {
		ts.T().Logf("Skipping test as `PACKER_ACC` is unset.")
		ts.T().Skip()
	}
}

func PackerCoreSuite(t *testing.T) (*PackerTestSuite, func()) {
	ts := &PackerTestSuite{}

	tempDir, err := os.MkdirTemp("", "packer-core-acc-test-")
	if err != nil {
		panic(fmt.Sprintf("failed to create temporary directory for compiled plugins: %s", err))
	}
	ts.pluginsDirectory = tempDir

	defer func() {
	}()

	packerPath := os.Getenv("PACKER_CUSTOM_PATH")
	if packerPath == "" {
		var err error
		t.Logf("Building test packer binary...")
		packerPath, err = BuildTestPacker(t)
		if err != nil {
			t.Fatalf("failed to build Packer binary: %s", err)
		}
	}
	ts.packerPath = packerPath
	t.Logf("Done")

	return ts, func() {
		err := os.RemoveAll(ts.pluginsDirectory)
		if err != nil {
			t.Logf("failed to cleanup directory %q: %s. This will need manual action", ts.pluginsDirectory, err)
		}

		if os.Getenv("PACKER_CUSTOM_PATH") != "" {
			return
		}

		err = os.Remove(ts.packerPath)
		if err != nil {
			t.Logf("failed to cleanup compiled packer binary %q: %s. This will need manual action", packerPath, err)
		}
	}
}
