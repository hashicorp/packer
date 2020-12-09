package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	pluginsdk "github.com/hashicorp/packer/packer-plugin-sdk/plugin"
)

func newConfig() Config {
	var conf Config
	conf.PluginMinPort = 10000
	conf.PluginMaxPort = 25000
	return conf
}

func TestDiscoverReturnsIfMagicCookieSet(t *testing.T) {
	config := newConfig()

	os.Setenv(pluginsdk.MagicCookieKey, pluginsdk.MagicCookieValue)
	defer os.Unsetenv(pluginsdk.MagicCookieKey)

	err := config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.builders) != 0 {
		t.Fatalf("Should not have tried to find builders")
	}
}

func TestEnvVarPackerPluginPath(t *testing.T) {
	// Create a temporary directory to store plugins in
	dir, _, cleanUpFunc, err := generateFakePlugins("custom_plugin_dir",
		[]string{"packer-provisioner-partyparrot"})
	if err != nil {
		t.Fatalf("Error creating fake custom plugins: %s", err)
	}

	defer cleanUpFunc()

	// Add temp dir to path.
	os.Setenv("PACKER_PLUGIN_PATH", dir)
	defer os.Unsetenv("PACKER_PLUGIN_PATH")

	config := newConfig()

	err = config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.provisioners) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if _, ok := config.provisioners["partyparrot"]; !ok {
		t.Fatalf("Should have found partyparrot provisioner.")
	}
}

func TestEnvVarPackerPluginPath_MultiplePaths(t *testing.T) {
	// Create a temporary directory to store plugins in
	dir, _, cleanUpFunc, err := generateFakePlugins("custom_plugin_dir",
		[]string{"packer-provisioner-partyparrot"})
	if err != nil {
		t.Fatalf("Error creating fake custom plugins: %s", err)
	}

	defer cleanUpFunc()

	pathsep := ":"
	if runtime.GOOS == "windows" {
		pathsep = ";"
	}

	// Create a second dir to look in that will be empty
	decoyDir, err := ioutil.TempDir("", "decoy")
	if err != nil {
		t.Fatalf("Failed to create a temporary test dir.")
	}
	defer os.Remove(decoyDir)

	pluginPath := dir + pathsep + decoyDir

	// Add temp dir to path.
	os.Setenv("PACKER_PLUGIN_PATH", pluginPath)
	defer os.Unsetenv("PACKER_PLUGIN_PATH")

	config := newConfig()

	err = config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.provisioners) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if _, ok := config.provisioners["partyparrot"]; !ok {
		t.Fatalf("Should have found partyparrot provisioner.")
	}
}

func generateFakePlugins(dirname string, pluginNames []string) (string, []string, func(), error) {
	dir, err := ioutil.TempDir("", dirname)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create temporary test directory: %v", err)
	}

	cleanUpFunc := func() {
		os.RemoveAll(dir)
	}

	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	plugins := make([]string, len(pluginNames))
	for i, plugin := range pluginNames {
		plug := filepath.Join(dir, plugin+suffix)
		plugins[i] = plug
		_, err := os.Create(plug)
		if err != nil {
			cleanUpFunc()
			return "", nil, nil, fmt.Errorf("failed to create temporary plugin file (%s): %v", plug, err)
		}
	}

	return dir, plugins, cleanUpFunc, nil
}
