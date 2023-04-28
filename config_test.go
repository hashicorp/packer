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

	"github.com/hashicorp/packer/packer"
)

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

	cfg := config{
		Plugins: &packer.PluginConfig{
			Builders:       packer.MapOfBuilder{},
			PostProcessors: packer.MapOfPostProcessor{},
			Provisioners:   packer.MapOfProvisioner{},
		},
	}

	if err := decodeConfig(strings.NewReader(packerConfigData), &cfg); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	cfg.LoadExternalComponentsFromConfig()

	if len(cfg.Plugins.Builders.List()) != 1 || !cfg.Plugins.Builders.Has("cloud-xyz") {
		t.Errorf("failed to load external builders; got %v as the resulting config", cfg.Plugins.Builders)
	}

	if len(cfg.Plugins.PostProcessors.List()) != 1 || !cfg.Plugins.PostProcessors.Has("noop") {
		t.Errorf("failed to load external post-processors; got %v as the resulting config", cfg.Plugins.PostProcessors)
	}

	if len(cfg.Plugins.Provisioners.List()) != 1 || !cfg.Plugins.Provisioners.Has("super-shell") {
		t.Errorf("failed to load external provisioners; got %v as the resulting config", cfg.Plugins.Provisioners)
	}

}

func TestLoadExternalComponentsFromConfig_onlyProvisioner(t *testing.T) {
	packerConfigData, cleanUpFunc, err := generateFakePackerConfigData()
	if err != nil {
		t.Fatalf("error encountered while creating fake Packer configuration data %v", err)
	}
	defer cleanUpFunc()

	cfg := config{
		Plugins: &packer.PluginConfig{
			Builders:       packer.MapOfBuilder{},
			PostProcessors: packer.MapOfPostProcessor{},
			Provisioners:   packer.MapOfProvisioner{},
		},
	}

	if err := decodeConfig(strings.NewReader(packerConfigData), &cfg); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	/* Let's clear out any custom Builders or PostProcessors that were part of the config.
	This step does not remove them from disk, it just removes them from of plugins Packer knows about.
	*/
	cfg.RawBuilders = nil
	cfg.RawPostProcessors = nil

	cfg.LoadExternalComponentsFromConfig()

	if len(cfg.Plugins.Builders.List()) != 0 {
		t.Errorf("loaded external builders when it wasn't supposed to; got %v as the resulting config", cfg.Plugins.Builders)
	}

	if len(cfg.Plugins.PostProcessors.List()) != 0 {
		t.Errorf("loaded external post-processors when it wasn't supposed to; got %v as the resulting config", cfg.Plugins.PostProcessors)
	}

	if len(cfg.Plugins.Provisioners.List()) != 1 || !cfg.Plugins.Provisioners.Has("super-shell") {
		t.Errorf("failed to load external provisioners; got %v as the resulting config", cfg.Plugins.Provisioners)
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

	cfg := config{
		Plugins: &packer.PluginConfig{
			Builders:       packer.MapOfBuilder{},
			PostProcessors: packer.MapOfPostProcessor{},
			Provisioners:   packer.MapOfProvisioner{},
		},
	}

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

/*
	generateFakePackerConfigData creates a collection of mock plugins along with a basic packerconfig.

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
