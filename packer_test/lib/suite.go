package lib

import (
	"fmt"
	"os"
	"sync"
	"testing"

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
	// compiledPlugins is the map of each compiled plugin to its path.
	//
	// This used to be global, but should be linked to the suite instead, as
	// we may have multiple suites that exist, each with its own repo of
	// plugins compiled for the purposes of the test, so as they all run
	// within the same process space, they should be separate instances.
	compiledPlugins sync.Map
}

func (ts *PackerTestSuite) buildPluginVersion(waitgroup *sync.WaitGroup, versionString string, t *testing.T) {
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		ts.CompilePlugin(t, versionString)
	}()
}

func (ts *PackerTestSuite) CompileTestPluginVersions(t *testing.T, versions ...string) {
	wg := &sync.WaitGroup{}

	for _, ver := range versions {
		ts.buildPluginVersion(wg, ver, t)
	}

	wg.Wait()
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

func InitBaseSuite(t *testing.T) (*PackerTestSuite, func()) {
	ts := &PackerTestSuite{
		compiledPlugins: sync.Map{},
	}

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
