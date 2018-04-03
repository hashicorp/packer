package fat

import "testing"

func TestGenerateShortName(t *testing.T) {
	// Test a basic one with no used
	result, err := generateShortName("foo.bar", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "FOO.BAR" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test long
	result, err = generateShortName("foobarbazblah.bar", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "FOOBAR~1.BAR" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test weird characters
	result, err = generateShortName("foo*b?r?baz.bar", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "FOO_B_~1.BAR" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test used
	result, err = generateShortName("foo.bar", []string{"foo.bar"})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "FOO~1.BAR" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test without a dot
	result, err = generateShortName("BAM", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "BAM" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test dotfile
	result, err = generateShortName(".big", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != ".BIG" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test valid extension
	result, err = generateShortName("proxy.psm", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "PROXY.PSM" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test long extension
	result, err = generateShortName("proxy.psm1", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "PROXY~1.PSM" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test short extension
	result, err = generateShortName("proxy.x", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "PROXY.X" {
		t.Fatalf("unexpected: %s", result)
	}

	// Test double shortname
	result, err = generateShortName("proxy.x", []string{"PROXY.X", "PROXY~1.X"})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	if result != "PROXY~2.X" {
		t.Fatalf("unexpected: %s", result)
	}
}

func TestShortNameEntryValue(t *testing.T) {
	// Test dot entry
	entryValue := shortNameEntryValue(".")
	expected := string([]byte{0x2E, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20})
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}

	// Test dotdot entry
	entryValue = shortNameEntryValue("..")
	expected = string([]byte{0x2E, 0x2E, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20})
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}

	// Test dotfile entry
	shortName, err := generateShortName(".big", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	entryValue = shortNameEntryValue(shortName)
	expected = "        BIG"
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}

	// Test entry value for a short filename without a period and file extension
	shortName, err = generateShortName("foo", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	entryValue = shortNameEntryValue(shortName)
	expected = "FOO        "
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}

	// Test entry value for a short filename with a period, but no file extension
	shortName, err = generateShortName("foo.", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	entryValue = shortNameEntryValue(shortName)
	expected = "FOO        "
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}

	// Test entry value for a short filename with a period and file extension
	shortName, err = generateShortName("foo.bar", []string{})
	if err != nil {
		t.Fatalf("err should be nil: %s", err)
	}

	entryValue = shortNameEntryValue(shortName)
	expected = "FOO     BAR"
	if entryValue != expected {
		t.Fatalf("expected %s, got %s", expected, entryValue)
	}
}
