package common

import (
	"testing"
)

func TestNilFloppies(t *testing.T) {
	c := FloppyConfig{}
	errs := c.Prepare(nil)
	if len(errs) != 0 {
		t.Fatal("nil floppies array should not fail")
	}

	if len(c.FloppyFiles) > 0 {
		t.Fatal("struct should not have floppy files")
	}
}

func TestEmptyArrayFloppies(t *testing.T) {
	c := FloppyConfig{
		FloppyFiles: make([]string, 0),
	}

	errs := c.Prepare(nil)
	if len(errs) != 0 {
		t.Fatal("empty floppies array should never fail")
	}

	if len(c.FloppyFiles) > 0 {
		t.Fatal("struct should not have floppy files")
	}
}

func TestExistingFloppyFile(t *testing.T) {
	c := FloppyConfig{
		FloppyFiles: []string{"floppy_config.go"},
	}

	errs := c.Prepare(nil)
	if len(errs) != 0 {
		t.Fatal("array with existing floppies should not fail")
	}
}

func TestNonExistingFloppyFile(t *testing.T) {
	c := FloppyConfig{
		FloppyFiles: []string{"floppy_config.foo"},
	}

	errs := c.Prepare(nil)
	if len(errs) == 0 {
		t.Fatal("array with non existing floppies should return errors")
	}
}

func TestMultiErrorFloppyFiles(t *testing.T) {
	c := FloppyConfig{
		FloppyFiles: []string{"floppy_config.foo", "floppy_config.go", "floppy_config.bar", "floppy_config_test.go", "floppy_config.baz"},
	}

	errs := c.Prepare(nil)
	if len(errs) == 0 {
		t.Fatal("array with non existing floppies should return errors")
	}

	expectedErrors := 3
	if count := len(errs); count != expectedErrors {
		t.Fatalf("array with %v non existing floppy should return %v errors but it is returning %v", expectedErrors, expectedErrors, count)
	}
}
