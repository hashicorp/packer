package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer/plugin"
)

func newConfig() config {
	var conf config
	conf.PluginMinPort = 10000
	conf.PluginMaxPort = 25000
	conf.Builders = packersdk.MapOfBuilder{}
	conf.PostProcessors = packersdk.MapOfPostProcessor{}
	conf.Provisioners = packersdk.MapOfProvisioner{}

	return conf
}
func TestDiscoverReturnsIfMagicCookieSet(t *testing.T) {
	config := newConfig()

	os.Setenv(plugin.MagicCookieKey, plugin.MagicCookieValue)
	defer os.Unsetenv(plugin.MagicCookieKey)

	err := config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.Builders) != 0 {
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

	if len(config.Provisioners) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if _, ok := config.Provisioners["partyparrot"]; !ok {
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

	if len(config.Provisioners) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if _, ok := config.Provisioners["partyparrot"]; !ok {
		t.Fatalf("Should have found partyparrot provisioner.")
	}
}

>>>>>>> move packer config constants next to the packer config
func TestDecodeConfig(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true,
		"provisioners": {
		    "super-shell": "packer-provisioner-super-shell"
		}
	}`

	var cfg config
	err := decodeConfig(strings.NewReader(packerConfig), &cfg)
	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	var expectedCfg config
	json.NewDecoder(strings.NewReader(packerConfig)).Decode(&expectedCfg)
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("failed to load custom configuration data; expected %v got %v", expectedCfg, cfg)
	}

}

func TestLoadExternalComponentsFromConfig(t *testing.T) {
	packerConfigData, cleanUpFunc, err := generateFakePackerConfigData()
	if err != nil {
		t.Fatalf("error encountered while creating fake Packer configuration data %v", err)
	}
	defer cleanUpFunc()

	var cfg config
	cfg.Builders = packersdk.MapOfBuilder{}
	cfg.PostProcessors = packersdk.MapOfPostProcessor{}
	cfg.Provisioners = packersdk.MapOfProvisioner{}

	if err := decodeConfig(strings.NewReader(packerConfigData), &cfg); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	cfg.LoadExternalComponentsFromConfig()

	if len(cfg.Builders) != 1 || !cfg.Builders.Has("cloud-xyz") {
		t.Errorf("failed to load external builders; got %v as the resulting config", cfg.Builders)
	}

	if len(cfg.PostProcessors) != 1 || !cfg.PostProcessors.Has("noop") {
		t.Errorf("failed to load external post-processors; got %v as the resulting config", cfg.PostProcessors)
	}

	if len(cfg.Provisioners) != 1 || !cfg.Provisioners.Has("super-shell") {
		t.Errorf("failed to load external provisioners; got %v as the resulting config", cfg.Provisioners)
	}

}

func TestLoadExternalComponentsFromConfig_onlyProvisioner(t *testing.T) {
	packerConfigData, cleanUpFunc, err := generateFakePackerConfigData()
	if err != nil {
		t.Fatalf("error encountered while creating fake Packer configuration data %v", err)
	}
	defer cleanUpFunc()

	var cfg config
	cfg.Provisioners = packersdk.MapOfProvisioner{}

	if err := decodeConfig(strings.NewReader(packerConfigData), &cfg); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	/* Let's clear out any custom Builders or PostProcessors that were part of the config.
	This step does not remove them from disk, it just removes them from of plugins Packer knows about.
	*/
	cfg.RawBuilders = nil
	cfg.RawPostProcessors = nil

	cfg.LoadExternalComponentsFromConfig()

	if len(cfg.Builders) != 0 {
		t.Errorf("loaded external builders when it wasn't supposed to; got %v as the resulting config", cfg.Builders)
	}

	if len(cfg.PostProcessors) != 0 {
		t.Errorf("loaded external post-processors when it wasn't supposed to; got %v as the resulting config", cfg.PostProcessors)
	}

	if len(cfg.Provisioners) != 1 || !cfg.Provisioners.Has("super-shell") {
		t.Errorf("failed to load external provisioners; got %v as the resulting config", cfg.Provisioners)
	}
}

func TestLoadSingleComponent(t *testing.T) {

	// .exe will work everyone for testing purpose, but mostly here to help Window's test runs.
	tmpFile, err := ioutil.TempFile(".", "packer-builder-*.exe")
	if err != nil {
		t.Fatalf("failed to create test file with error: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	tt := []struct {
		pluginPath    string
		errorExpected bool
	}{
		{pluginPath: tmpFile.Name(), errorExpected: false},
		{pluginPath: "./non-existing-file", errorExpected: true},
	}

	var cfg config
	cfg.Builders = packersdk.MapOfBuilder{}
	cfg.PostProcessors = packersdk.MapOfPostProcessor{}
	cfg.Provisioners = packersdk.MapOfProvisioner{}

	for _, tc := range tt {
		tc := tc
		_, err := cfg.loadSingleComponent(tc.pluginPath)
		if tc.errorExpected && err == nil {
			t.Errorf("expected loadSingleComponent(%s) to error but it didn't", tc.pluginPath)
			continue
		}

		if err != nil && !tc.errorExpected {
			t.Errorf("expected loadSingleComponent(%s) to load properly but got an error: %v", tc.pluginPath, err)
		}
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

/* generateFakePackerConfigData creates a collection of mock plugins along with a basic packerconfig.
The return packerConfigData is a valid packerconfig file that can be used for configuring external plugins, cleanUpFunc is a function that should be called for cleaning up any generated mock data.
This function will only clean up if there is an error, on successful runs the caller
is responsible for cleaning up the data via cleanUpFunc().
*/
func generateFakePackerConfigData() (packerConfigData string, cleanUpFunc func(), err error) {
	_, plugins, cleanUpFunc, err := generateFakePlugins("random-testdata",
		[]string{"packer-builder-cloud-xyz",
			"packer-provisioner-super-shell",
			"packer-post-processor-noop"})

	if err != nil {
		cleanUpFunc()
		return "", nil, err
	}

	packerConfigData = fmt.Sprintf(`
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true,
		"builders": {
			"cloud-xyz": %q
		},
		"provisioners": {
			"super-shell": %q
		},
		"post-processors": {
			"noop": %q
		}
	}`, plugins[0], plugins[1], plugins[2])

	return
}
