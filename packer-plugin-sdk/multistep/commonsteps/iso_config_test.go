// +build !windows

package commonsteps

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testISOConfig() ISOConfig {
	return ISOConfig{
		ISOChecksum:     "md5:0B0F137F17AC10944716020B018F8126",
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
	i.ISOChecksum = "0b0F137F17AC10944716020B018F8126"
	warns, err = i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

}

func TestISOConfigPrepare_ISOChecksumURLBad(t *testing.T) {
	// Test that we won't try to read an iso into memory because of a user
	// error
	i := testISOConfig()
	i.ISOChecksum = "file:///not_a_checksum.iso"
	_, err := i.Prepare(nil)
	if err == nil {
		t.Fatalf("should have error because iso is bad filetype: %s", err)
	}
}

func TestISOConfigPrepare_ISOChecksumType(t *testing.T) {
	i := testISOConfig()

	// Test none
	i = testISOConfig()
	i.ISOChecksum = "none"
	warns, err := i.Prepare(nil)
	if len(warns) == 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
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
	ts := httpTestModule("root")
	defer ts.Close()
	i = testISOConfig()
	i.RawSingleISOUrl = ""
	i.ISOChecksum = "file:" + ts.URL + "/basic.txt"
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

func TestISOConfigPrepare_ISOChecksumURLMyTest(t *testing.T) {
	httpChecksums := httpTestModule("root")
	defer httpChecksums.Close()
	i := ISOConfig{
		ISOChecksum: "file:" + httpChecksums.URL + "/subfolder.sum",
		ISOUrls:     []string{"http://hashicorp.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso"},
	}

	// Test ISOChecksum overrides url
	warns, err := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("Bad: should not have warnings")
	}
	if len(err) > 0 {
		t.Fatalf("Bad; should not have errored.")
	}
}

func TestISOConfigPrepare_ISOChecksumLocalFile(t *testing.T) {
	// Creates checksum file in local dir
	p := filepath.Join(fixtureDir, "root/subfolder.sum")
	source, err := os.Open(p)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer source.Close()
	destination, err := os.Create("local.sum")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Remove("local.sum")
	defer destination.Close()
	if _, err := io.Copy(destination, source); err != nil {
		t.Fatalf(err.Error())
	}

	i := ISOConfig{
		ISOChecksum: "file:./local.sum",
		ISOUrls:     []string{"http://hashicorp.com/ubuntu/dists/bionic-updates/main/installer-amd64/current/images/netboot/mini.iso"},
	}

	warns, errs := i.Prepare(nil)
	if len(warns) > 0 {
		t.Fatalf("Bad: should not have warnings")
	}
	if len(errs) > 0 {
		t.Fatalf("Bad; should not have errored. %v", errs)
	}
}

const fixtureDir = "./test-fixtures"

func httpTestModule(n string) *httptest.Server {
	p := filepath.Join(fixtureDir, n)
	p, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}

	return httptest.NewServer(http.FileServer(http.Dir(p)))
}
