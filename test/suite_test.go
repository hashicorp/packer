package test

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
		BuildSimplePlugin(NewPluginBuildConfig(versionString), t)
		waitgroup.Done()
	}()
}

func (ts *PackerTestSuite) buildPluginBinaries(t *testing.T) {
	wg := &sync.WaitGroup{}

	buildPluginVersion(wg, "1.0.0", t)
	buildPluginVersion(wg, "1.0.1-dev", t)
	buildPluginVersion(wg, "1.0.1", t)
	buildPluginVersion(wg, "1.0.9", t)
	buildPluginVersion(wg, "1.0.10", t)

	wg.Wait()
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

	t.Logf("Building test packer binary...")
	packerPath, err := BuildTestPacker(t)
	if err != nil {
		t.Fatalf("failed to build Packer binary: %s", err)
	}
	ts.packerPath = packerPath
	t.Logf("Done")

	defer func() {
		err := os.Remove(ts.packerPath)
		if err != nil {
			t.Logf("failed to cleanup compiled packer binary %q: %s. This will need manual aciton", packerPath, err)
		}
	}()

	suite.Run(t, ts)
}
