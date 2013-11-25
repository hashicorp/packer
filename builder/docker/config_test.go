package docker

import (
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"export_path": "foo",
		"image":       "bar",
	}
}

func testConfigStruct(t *testing.T) *Config {
	c, warns, errs := NewConfig(testConfig())
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return c
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
	raw := testConfig()

	// No export path
	delete(raw, "export_path")
	_, warns, errs := NewConfig(raw)
	testConfigErr(t, warns, errs)

	// Good export path
	raw["export_path"] = "good"
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_image(t *testing.T) {
	raw := testConfig()

	// No image
	delete(raw, "image")
	_, warns, errs := NewConfig(raw)
	testConfigErr(t, warns, errs)

	// Good image
	raw["image"] = "path"
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_pull(t *testing.T) {
	raw := testConfig()

	// No pull set
	delete(raw, "pull")
	c, warns, errs := NewConfig(raw)
	testConfigOk(t, warns, errs)
	if !c.Pull {
		t.Fatal("should pull by default")
	}

	// Pull set
	raw["pull"] = false
	c, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)
	if c.Pull {
		t.Fatal("should not pull")
	}
}
