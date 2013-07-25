package chefSolo

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
	// "inline": []interface{}{"foo", "bar"},
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

// Cookbook paths
//////////////////

func TestProvisionerPrepare_DefaultCookbookPathIsUsed(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Errorf("expected error to be thrown for unavailable cookbook path")
	}

	if len(p.config.CookbooksPaths) != 1 || p.config.CookbooksPaths[0] != DefaultCookbooksPath {
		t.Errorf("unexpected default cookbook path: %s", p.config.CookbooksPaths)
	}
}

func TestProvisionerPrepare_GivenCookbookPathsAreAddedToConfig(t *testing.T) {
	var p Provisioner

	path1, err := ioutil.TempDir("", "cookbooks_one")
	if err != nil {
		t.Errorf("err: %s", err)
	}

	path2, err := ioutil.TempDir("", "cookbooks_two")
	if err != nil {
		t.Errorf("err: %s", err)
	}

	defer os.Remove(path1)
	defer os.Remove(path2)

	config := testConfig()
	config["cookbooks_paths"] = []string{path1, path2}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	if len(p.config.CookbooksPaths) != 2 || p.config.CookbooksPaths[0] != path1 || p.config.CookbooksPaths[1] != path2 {
		t.Errorf("unexpected default cookbook path: %s", p.config.CookbooksPaths)
	}
}
