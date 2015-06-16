package file

import (
	"strings"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"source":  "src.txt",
		"target":  "dst.txt",
		"content": "Hello, world!",
	}
}

func TestContentSourceConflict(t *testing.T) {
	raw := testConfig()

	_, _, errs := NewConfig(raw)
	if !strings.Contains(errs.Error(), ErrContentSourceConflict.Error()) {
		t.Errorf("Expected config error: %s", ErrContentSourceConflict.Error())
	}
}

func TestNoFilename(t *testing.T) {
	raw := testConfig()

	delete(raw, "filename")
	_, _, errs := NewConfig(raw)
	if errs == nil {
		t.Errorf("Expected config error: %s", ErrTargetRequired.Error())
	}
}

func TestNoContent(t *testing.T) {
	raw := testConfig()

	delete(raw, "content")
	delete(raw, "source")
	_, warns, _ := NewConfig(raw)

	if len(warns) == 0 {
		t.Error("Expected config warning without any content")
	}
}
