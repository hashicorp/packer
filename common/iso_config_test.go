// +build !windows

package common

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func testISOConfig() ISOConfig {
	return ISOConfig{
		ISOChecksum:     "foo",
		ISOChecksumURL:  "",
		ISOChecksumType: "md5",
		RawSingleISOUrl: "http://www.packer.io/the-OS.iso",
	}
}

var cs_bsd_style = `
MD5 (other.iso) = bAr
MD5 (the-OS.iso) = baZ
`

var cs_bsd_style_subdir = `
MD5 (other.iso) = bAr
MD5 (./subdir/the-OS.iso) = baZ
`

var cs_gnu_style = `
bAr0 *the-OS.iso
baZ0  other.iso
`

var cs_gnu_style_subdir = `
bAr0 *./subdir/the-OS.iso
baZ0  other.iso
`

var cs_bsd_style_no_newline = `
MD5 (other.iso) = bAr
MD5 (the-OS.iso) = baZ`

var cs_gnu_style_no_newline = `
bAr0 *the-OS.iso
baZ0  other.iso`

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

}

func TestISOConfigPrepare_ISOChecksumURLBad(t *testing.T) {
	i := testISOConfig()
	i.ISOChecksumURL = "file:///not_read"
	i.ISOChecksum = "shouldoverride"

	// Test ISOChecksum overrides url
	warns, err := i.Prepare(nil)
	if len(warns) != 1 {
		t.Fatalf("Bad: should have warned because both checksum and " +
			"checksumURL are set.")
	}
	if len(err) > 0 {
		t.Fatalf("Bad; should have warned but not errored.")
	}

	// Test that we won't try to read an iso into memory because of a user
	// error
	i = testISOConfig()
	i.ISOChecksumURL = "file:///not_read.iso"
	i.ISOChecksum = ""
	warns, err = i.Prepare(nil)
	if err == nil {
		t.Fatalf("should have error because iso is bad filetype: %s", err)
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
	if err != nil {
		t.Fatalf("should not have error: %s", err)
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

	// Test iso_url not set but checksum url is
	ts := httptest.NewServer(http.FileServer(http.Dir("./test-fixtures/root")))
	defer ts.Close()
	i = testISOConfig()
	i.RawSingleISOUrl = ""
	i.ISOChecksum = ""
	i.ISOChecksumURL = ts.URL + "/basic.txt"
	// ISOConfig.Prepare() returns a slice of errors
	var errs []error
	warns, errs = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("expected no warnings, got:%v", warns)
	}
	if len(errs) < 1 || err[0] == nil {
		t.Fatalf("expected a populated error slice, got: %v", errs)
	}

	// Test iso_url set
	i = testISOConfig()
	i.RawSingleISOUrl = "http://www.packer.io/the-OS.iso"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected := []string{"http://www.packer.io/the-OS.iso"}
	if !reflect.DeepEqual(i.ISOUrls, expected) {
		t.Fatalf("bad: %#v", i.ISOUrls)
	}

	// Test both set
	i = testISOConfig()
	i.RawSingleISOUrl = "http://www.packer.io/the-OS.iso"
	i.ISOUrls = []string{"http://www.packer.io/the-OS.iso"}
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
		"http://www.packer.io/the-OS.iso",
		"http://www.hashicorp.com/the-OS.iso",
	}

	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Errorf("should not have error: %s", err)
	}

	expected = []string{
		"http://www.packer.io/the-OS.iso",
		"http://www.hashicorp.com/the-OS.iso",
	}
	if !reflect.DeepEqual(i.ISOUrls, expected) {
		t.Fatalf("bad: %#v", i.ISOUrls)
	}
}

func TestISOConfigPrepare_TargetExtension(t *testing.T) {
	i := testISOConfig()

	// Test the default value
	warns, err := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.TargetExtension != "iso" {
		t.Fatalf("should've found \"iso\" got: %s", i.TargetExtension)
	}

	// Test the lowercased value
	i = testISOConfig()
	i.TargetExtension = "DMG"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.TargetExtension != "dmg" {
		t.Fatalf("should've lowercased: %s", i.TargetExtension)
	}
}
