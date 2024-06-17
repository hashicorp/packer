package packer_test

import (
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
}

func buildPluginVersion(waitgroup *sync.WaitGroup, versionString string, t *testing.T) {
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		BuildSimplePlugin(versionString, t)
	}()
}

func (ts *PackerTestSuite) buildPluginBinaries(t *testing.T) {
	wg := &sync.WaitGroup{}

	buildPluginVersion(wg, "1.0.0", t)
	buildPluginVersion(wg, "1.0.0+metadata", t)
	buildPluginVersion(wg, "1.0.1-alpha1", t)
	buildPluginVersion(wg, "1.0.9", t)
	buildPluginVersion(wg, "1.0.10", t)

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

func Test_PackerCoreSuite(t *testing.T) {
	ts := &PackerTestSuite{}

	pluginsDirectory := PluginBinaryDir()
	defer func() {
		err := os.RemoveAll(pluginsDirectory)
		if err != nil {
			t.Logf("failed to cleanup directory %q: %s. This will need manual action", pluginsDirectory, err)
		}
	}()

	ts.pluginsDirectory = pluginsDirectory
	ts.buildPluginBinaries(t)

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

	defer func() {
		if os.Getenv("PACKER_CUSTOM_PATH") != "" {
			return
		}

		err := os.Remove(ts.packerPath)
		if err != nil {
			t.Logf("failed to cleanup compiled packer binary %q: %s. This will need manual action", packerPath, err)
		}
	}()

	suite.Run(t, ts)
}
