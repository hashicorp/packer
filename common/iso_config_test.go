package common

import (
	"reflect"
	"testing"
)

func testISOConfig() ISOConfig {
	return ISOConfig{
		ISOChecksum:     "foo",
		ISOChecksumType: "md5",
		RawSingleISOUrl: "http://www.packer.io",
	}
}

func TestISOConfigPrepare_ISOChecksum(t *testing.T) {
	i := testISOConfig()

	// Test bad
	i.ISOChecksum = ""
	warns, err := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	i = testISOConfig()
	i.ISOChecksum = "FOo"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksum != "foo" {
		t.Fatalf("should've lowercased: %s", i.ISOChecksum)
	}
}

func TestISOConfigPrepare_ISOChecksumType(t *testing.T) {
	i := testISOConfig()

	// Test bad
	i.ISOChecksumType = ""
	warns, err := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	i = testISOConfig()
	i.ISOChecksumType = "mD5"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksumType != "md5" {
		t.Fatalf("should've lowercased: %s", i.ISOChecksumType)
	}

	// Test unknown
	i = testISOConfig()
	i.ISOChecksumType = "fake"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test none
	i = testISOConfig()
	i.ISOChecksumType = "none"
	warns, err = i.Prepare(nil)
	if len(warns) == 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksumType != "none" {
		t.Fatalf("should've lowercased: %s", i.ISOChecksumType)
	}
}

func TestISOConfigPrepare_ISOUrl(t *testing.T) {
	i := testISOConfig()

	// Test both empty
	i.RawSingleISOUrl = ""
	i.ISOUrls = []string{}
	warns, err := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test iso_url set
	i = testISOConfig()
	i.RawSingleISOUrl = "http://www.packer.io"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected := []string{"http://www.packer.io"}
	if !reflect.DeepEqual(i.ISOUrls, expected) {
		t.Fatalf("bad: %#v", i.ISOUrls)
	}

	// Test both set
	i = testISOConfig()
	i.RawSingleISOUrl = "http://www.packer.io"
	i.ISOUrls = []string{"http://www.packer.io"}
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test just iso_urls set
	i = testISOConfig()
	i.RawSingleISOUrl = ""
	i.ISOUrls = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}

	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected = []string{
		"http://www.packer.io",
		"http://www.hashicorp.com",
	}
	if !reflect.DeepEqual(i.ISOUrls, expected) {
		t.Fatalf("bad: %#v", i.ISOUrls)
	}
}
