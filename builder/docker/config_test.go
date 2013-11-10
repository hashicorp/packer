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

func TestConfigPrepare_exportPath(t *testing.T) {
	raw := testConfig()

	// No export path
	delete(raw, "export_path")
	_, warns, errs := NewConfig(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if errs == nil {
		t.Fatal("should error")
	}

	// Good export path
	raw["export_path"] = "good"
	_, warns, errs = NewConfig(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if errs != nil {
		t.Fatalf("bad: %s", errs)
	}
}

func TestConfigPrepare_image(t *testing.T) {
	raw := testConfig()

	// No image
	delete(raw, "image")
	_, warns, errs := NewConfig(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if errs == nil {
		t.Fatal("should error")
	}

	// Good image
	raw["image"] = "path"
	_, warns, errs = NewConfig(raw)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if errs != nil {
		t.Fatalf("bad: %s", errs)
	}
}
