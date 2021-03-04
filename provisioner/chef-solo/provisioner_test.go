package chefsolo

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_chefEnvironment(t *testing.T) {
	var p Provisioner

	config := testConfig()
	config["chef_environment"] = "some-env"

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ChefEnvironment != "some-env" {
		t.Fatalf("unexpected: %#v", p.config.ChefEnvironment)
	}
}
func TestProvisionerPrepare_chefLicense(t *testing.T) {
	var p Provisioner

	// Test not set
	config := testConfig()
	err := p.Prepare(config)
	if err != nil {
		t.Fatal("should error")
	}

	if p.config.ChefLicense != "accept-silent" {
		t.Fatalf("unexpected: %#v", p.config.ChefLicense)
	}

	// Test set
	config = testConfig()
	config["chef_license"] = "accept"
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ChefLicense != "accept" {
		t.Fatalf("unexpected: %#v", p.config.ChefLicense)
	}

	// Test set skipInstall true
	config = testConfig()
	config["skip_install"] = true
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ChefLicense != "" {
		t.Fatalf("unexpected: %#v", "empty string")
	}

	// Test set installCommand true
	config = testConfig()
	config["install_command"] = "install chef"
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ChefLicense != "" {
		t.Fatalf("unexpected: %#v", "empty string")
	}
}

func TestProvisionerPrepare_configTemplate(t *testing.T) {
	var err error
	var p Provisioner

	// Test no config template
	config := testConfig()
	delete(config, "config_template")
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a file
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())

	config = testConfig()
	config["config_template"] = tf.Name()
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a directory
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	config = testConfig()
	config["config_template"] = td
	p = Provisioner{}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have err")
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

	rolesPath, err := ioutil.TempDir("", "roles")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	dataBagsPath, err := ioutil.TempDir("", "data_bags")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer os.Remove(path1)
	defer os.Remove(path2)
	defer os.Remove(rolesPath)
	defer os.Remove(dataBagsPath)

	config := testConfig()
	config["cookbook_paths"] = []string{path1, path2}
	config["roles_path"] = rolesPath
	config["data_bags_path"] = dataBagsPath

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

	if p.config.RolesPath != rolesPath {
		t.Fatalf("unexpected: %#v", p.config.RolesPath)
	}

	if p.config.DataBagsPath != dataBagsPath {
		t.Fatalf("unexpected: %#v", p.config.DataBagsPath)
	}
}

func TestProvisionerPrepare_dataBagsPath(t *testing.T) {
	var p Provisioner

	dataBagsPath, err := ioutil.TempDir("", "data_bags")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(dataBagsPath)

	config := testConfig()
	config["data_bags_path"] = dataBagsPath

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.DataBagsPath != dataBagsPath {
		t.Fatalf("unexpected: %#v", p.config.DataBagsPath)
	}
}

func TestProvisionerPrepare_encryptedDataBagSecretPath(t *testing.T) {
	var err error
	var p Provisioner

	// Test no config template
	config := testConfig()
	delete(config, "encrypted_data_bag_secret_path")
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a file
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())

	config = testConfig()
	config["encrypted_data_bag_secret_path"] = tf.Name()
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a directory
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	config = testConfig()
	config["encrypted_data_bag_secret_path"] = td
	p = Provisioner{}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestProvisionerPrepare_environmentsPath(t *testing.T) {
	var p Provisioner

	environmentsPath, err := ioutil.TempDir("", "environments")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(environmentsPath)

	config := testConfig()
	config["environments_path"] = environmentsPath

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.EnvironmentsPath != environmentsPath {
		t.Fatalf("unexpected: %#v", p.config.EnvironmentsPath)
	}
}

func TestProvisionerPrepare_rolesPath(t *testing.T) {
	var p Provisioner

	rolesPath, err := ioutil.TempDir("", "roles")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(rolesPath)

	config := testConfig()
	config["roles_path"] = rolesPath

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.RolesPath != rolesPath {
		t.Fatalf("unexpected: %#v", p.config.RolesPath)
	}
}

func TestProvisionerPrepare_json(t *testing.T) {
	config := testConfig()
	config["json"] = map[string]interface{}{
		"foo": "{{ user `foo` }}",
	}

	config[common.UserVariablesConfigKey] = map[string]string{
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

func TestProvisionerPrepare_jsonNested(t *testing.T) {
	config := testConfig()
	config["json"] = map[string]interface{}{
		"foo": map[interface{}]interface{}{
			"bar": []uint8("baz"),
		},

		"bar": []interface{}{
			"foo",

			map[interface{}]interface{}{
				"bar": "baz",
			},
		},

		"bFalse": false,
		"bTrue":  true,
		"bNil":   nil,
		"bStr":   []uint8("bar"),

		"bInt":   1,
		"bFloat": 4.5,
	}

	var p Provisioner
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	fooMap := p.config.Json["foo"].(map[string]interface{})
	if fooMap["bar"] != "baz" {
		t.Fatalf("nope: %#v", fooMap["bar"])
	}
	if p.config.Json["bStr"] != "bar" {
		t.Fatalf("nope: %#v", fooMap["bar"])
	}
}

func TestProvisionerPrepare_jsonstring(t *testing.T) {
	config := testConfig()
	config["json_string"] = `{
			"foo": {
				"bar": "baz"
			},
			"bar": {
				"bar": "baz"
			},	
			"bFalse": false,
			"bTrue": true,
			"bStr": "bar",
			"bNil": null,
			"bInt": 1,
			"bFloat": 4.5
		}`

	var p Provisioner
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	fooMap := p.config.Json["foo"].(map[string]interface{})
	if fooMap["bar"] != "baz" {
		t.Fatalf("nope: %#v", fooMap["bar"])
	}
	if p.config.Json["bStr"] != "bar" {
		t.Fatalf("nope: %#v", fooMap["bar"])
	}
}
