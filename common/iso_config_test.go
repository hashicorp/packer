package common

import (
	"fmt"
	"io/ioutil"
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

var cs_gnu_style = `
bAr0 *the-OS.iso
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

	if i.ISOChecksum != "foo" {
		t.Fatalf("should've lowercased: %s", i.ISOChecksum)
	}
}

func TestISOConfigPrepare_ISOChecksumURL(t *testing.T) {
	i := testISOConfig()
	i.ISOChecksumURL = "file:///not_read"

	// Test ISOChecksum overrides url
	warns, err := i.Prepare(nil)
	if len(warns) > 0 && len(err) > 0 {
		t.Fatalf("bad: %#v, %#v", warns, err)
	}

	// Test good - ISOChecksumURL BSD style
	i = testISOConfig()
	i.ISOChecksum = ""
	cs_file, _ := ioutil.TempFile("", "packer-test-")
	ioutil.WriteFile(cs_file.Name(), []byte(cs_bsd_style), 0666)
	i.ISOChecksumURL = fmt.Sprintf("file://%s", cs_file.Name())
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksum != "baz" {
		t.Fatalf("should've found \"baz\" got: %s", i.ISOChecksum)
	}

	// Test good - ISOChecksumURL GNU style
	i = testISOConfig()
	i.ISOChecksum = ""
	cs_file, _ = ioutil.TempFile("", "packer-test-")
	ioutil.WriteFile(cs_file.Name(), []byte(cs_gnu_style), 0666)
	i.ISOChecksumURL = fmt.Sprintf("file://%s", cs_file.Name())
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksum != "bar0" {
		t.Fatalf("should've found \"bar0\" got: %s", i.ISOChecksum)
	}

	// Test good - ISOChecksumURL BSD style no newline
	i = testISOConfig()
	i.ISOChecksum = ""
	cs_file, _ = ioutil.TempFile("", "packer-test-")
	ioutil.WriteFile(cs_file.Name(), []byte(cs_bsd_style_no_newline), 0666)
	i.ISOChecksumURL = fmt.Sprintf("file://%s", cs_file.Name())
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksum != "baz" {
		t.Fatalf("should've found \"baz\" got: %s", i.ISOChecksum)
	}

	// Test good - ISOChecksumURL GNU style no newline
	i = testISOConfig()
	i.ISOChecksum = ""
	cs_file, _ = ioutil.TempFile("", "packer-test-")
	ioutil.WriteFile(cs_file.Name(), []byte(cs_gnu_style_no_newline), 0666)
	i.ISOChecksumURL = fmt.Sprintf("file://%s", cs_file.Name())
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if i.ISOChecksum != "bar0" {
		t.Fatalf("should've found \"bar0\" got: %s", i.ISOChecksum)
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
