package file

import (
	"fmt"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"filename": "test.txt",
		"contents": "Hello, world!",
	}
}

func TestNoFilename(t *testing.T) {
	raw := testConfig()

	delete(raw, "filename")
	_, _, errs := NewConfig(raw)
	if errs == nil {
		t.Error("Expected config to error without a filename")
	}
}

func TestNoContent(t *testing.T) {
	raw := testConfig()

	delete(raw, "contents")
	_, warns, _ := NewConfig(raw)
	fmt.Println(len(warns))
	fmt.Printf("%#v\n", warns)
	if len(warns) == 0 {
		t.Error("Expected config to warn without any content")
	}
}
