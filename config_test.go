package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestDecodeConfig_basic(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true
	}`

	var cfg config
	err := DecodeConfig(strings.NewReader(packerConfig), &cfg)
	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	var expectedCfg config
	json.NewDecoder(strings.NewReader(packerConfig)).Decode(&expectedCfg)
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("failed to load custom configuration data; expected %v got %v", expectedCfg, cfg)
	}

}

func TestDecodeConfig_plugins(t *testing.T) {

	dir, err := ioutil.TempDir("", "random-test-dir")
	if err != nil {
		t.Fatalf("failed to create temporary test directory: %v", err)
	}
	defer os.RemoveAll(dir)

	plugins := [...]string{
		filepath.Join(dir, "packer-builder-comment"),
		filepath.Join(dir, "packer-provisioner-comment"),
		filepath.Join(dir, "packer-post-processor-comment"),
	}
	for _, plugin := range plugins {
		_, err := os.Create(plugin)
		if err != nil {
			t.Fatalf("failed to create temporary plugin file (%s): %v", plugin, err)
		}
	}

	packerConfig := fmt.Sprintf(`
	{
		"builders": {
			"comment": %q
		},
		"provisioners": {
			"comment": %q
		},
		"post-processors": {
			"comment": %q
		}
	}`, plugins[0], plugins[1], plugins[2])

	var cfg config
	cfg.Builders = packer.MapOfBuilder{}
	cfg.PostProcessors = packer.MapOfPostProcessor{}
	cfg.Provisioners = packer.MapOfProvisioner{}
	err = DecodeConfig(strings.NewReader(packerConfig), &cfg)

	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	if len(cfg.RawBuilders) != 1 || !cfg.Builders.Has("comment") {
		t.Errorf("DecodeConfig failed to load external builders; got %#v as resulting config", cfg)
	}

	if len(cfg.RawProvisioners) != 1 || !cfg.Provisioners.Has("comment") {
		t.Errorf("DecodeConfig failed to load external provisioners; got %#v as resulting config", cfg)
	}

	if len(cfg.RawPostProcessors) != 1 || !cfg.PostProcessors.Has("comment") {
		t.Errorf("DecodeConfig failed to load external post-processors; got %#v as resulting config", cfg)
	}

}
