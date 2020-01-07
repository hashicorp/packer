package docker

import (
	"io/ioutil"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"export_path": "foo",
		"image":       "bar",
	}
}

func testConfigStruct(t *testing.T) *Config {
	var c Config
	warns, errs := c.Prepare(testConfig())
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return &c
}

func testConfigErr(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestConfigPrepare_exportPath(t *testing.T) {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	raw := testConfig()

	// No export path. This is invalid. Previously this would not error during
	// validation and as a result the failure would happen at build time.
	delete(raw, "export_path")
	var c Config
	warns, errs := c.Prepare(raw)
	testConfigErr(t, warns, errs)

	// Good export path
	raw["export_path"] = "good"
	warns, errs = c.Prepare(raw)
	testConfigOk(t, warns, errs)

	// Bad export path (directory)
	raw["export_path"] = td
	warns, errs = c.Prepare(raw)
	testConfigErr(t, warns, errs)
}

func TestConfigPrepare_exportPathAndCommit(t *testing.T) {
	raw := testConfig()

	// Export but no commit (explicit default)
	raw["commit"] = false
	warns, errs := (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)

	// Commit AND export specified (invalid)
	raw["commit"] = true
	warns, errs = (&Config{}).Prepare(raw)
	testConfigErr(t, warns, errs)

	// Commit but no export
	delete(raw, "export_path")
	warns, errs = (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_exportDiscard(t *testing.T) {
	raw := testConfig()

	// Export but no discard (explicit default)
	raw["discard"] = false
	warns, errs := (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)

	// Discard AND export (invalid)
	raw["discard"] = true
	warns, errs = (&Config{}).Prepare(raw)
	testConfigErr(t, warns, errs)

	// Discard but no export
	raw["discard"] = true
	delete(raw, "export_path")
	warns, errs = (&Config{}).Prepare(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_image(t *testing.T) {
	raw := testConfig()

	// No image
	delete(raw, "image")
	var c Config
	warns, errs := c.Prepare(raw)
	testConfigErr(t, warns, errs)

	// Good image
	raw["image"] = "path"
	warns, errs = c.Prepare(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_pull(t *testing.T) {
	raw := testConfig()

	// No pull set
	delete(raw, "pull")
	var c Config
	warns, errs := c.Prepare(raw)
	testConfigOk(t, warns, errs)
	if !c.Pull {
		t.Fatal("should pull by default")
	}

	// Pull set
	raw["pull"] = false
	warns, errs = c.Prepare(raw)
	testConfigOk(t, warns, errs)
	if c.Pull {
		t.Fatal("should not pull")
	}
}
