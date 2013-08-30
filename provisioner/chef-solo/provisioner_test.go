package chefsolo

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_cookbookPaths(t *testing.T) {
	var p Provisioner

	path1, err := ioutil.TempDir("", "cookbooks_one")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	path2, err := ioutil.TempDir("", "cookbooks_two")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer os.Remove(path1)
	defer os.Remove(path2)

	config := testConfig()
	config["cookbook_paths"] = []string{path1, path2}

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(p.config.CookbookPaths) != 2 {
		t.Fatalf("unexpected: %#v", p.config.CookbookPaths)
	}

	if p.config.CookbookPaths[0] != path1 || p.config.CookbookPaths[1] != path2 {
		t.Fatalf("unexpected: %#v", p.config.CookbookPaths)
	}
}

func TestProvisionerPrepare_json(t *testing.T) {
	config := testConfig()
	config["json"] = map[string]interface{}{
		"foo": "{{ user `foo` }}",
	}

	config[packer.UserVariablesConfigKey] = map[string]string{
		"foo": `"bar\baz"`,
	}

	var p Provisioner
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.Json["foo"] != `"bar\baz"` {
		t.Fatalf("bad: %#v", p.config.Json)
	}
}
