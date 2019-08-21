package vagrantcloud

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testGoodConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_token":        "foo",
		"version_description": "bar",
		"box_tag":             "hashicorp/precise64",
		"version":             "0.5",
	}
}

func testBadConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_token":        "foo",
		"box_tag":             "baz",
		"version_description": "bar",
	}
}

func newSecureServer(token string, handler http.HandlerFunc) *httptest.Server {
	token = fmt.Sprintf("Bearer %s", token)
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("authorization") != token {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if handler != nil {
			handler(rw, req)
		}
	}))
}

func newSelfSignedSslServer(token string, handler http.HandlerFunc) *httptest.Server {
	token = fmt.Sprintf("Bearer %s", token)
	return httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("authorization") != token {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if handler != nil {
			handler(rw, req)
		}
	}))
}

func TestPostProcessor_Insecure_Ssl(t *testing.T) {
	var p PostProcessor
	server := newSelfSignedSslServer("foo", nil)
	defer server.Close()

	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL
	config["insecure_skip_tls_verify"] = true
	if err := p.Configure(config); err != nil {
		t.Fatalf("Expected TLS to skip certificate validation: %s", err)
	}
}

func TestPostProcessor_Configure_fromVagrantEnv(t *testing.T) {
	var p PostProcessor
	config := testGoodConfig()
	server := newSecureServer("bar", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	config["access_token"] = ""
	os.Setenv("VAGRANT_CLOUD_TOKEN", "bar")
	defer func() {
		os.Setenv("VAGRANT_CLOUD_TOKEN", "")
	}()

	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.AccessToken != "bar" {
		t.Fatalf("Expected to get token from VAGRANT_CLOUD_TOKEN env var. Got '%s' instead",
			p.config.AccessToken)
	}
}

func TestPostProcessor_Configure_fromAtlasEnv(t *testing.T) {
	var p PostProcessor
	config := testGoodConfig()
	config["access_token"] = ""
	server := newSecureServer("foo", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	os.Setenv("ATLAS_TOKEN", "foo")
	defer func() {
		os.Setenv("ATLAS_TOKEN", "")
	}()

	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.AccessToken != "foo" {
		t.Fatalf("Expected to get token from ATLAS_TOKEN env var. Got '%s' instead",
			p.config.AccessToken)
	}

	if !p.warnAtlasToken {
		t.Fatal("Expected warn flag to be set when getting token from atlas env var.")
	}
}

func TestPostProcessor_Configure_Good(t *testing.T) {
	config := testGoodConfig()
	server := newSecureServer("foo", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	var p PostProcessor
	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_Configure_Bad(t *testing.T) {
	config := testBadConfig()
	server := newSecureServer("foo", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	var p PostProcessor
	if err := p.Configure(config); err == nil {
		t.Fatalf("should have err")
	}
}

func TestPostProcessor_PostProcess_checkArtifactType(t *testing.T) {
	artifact := &packer.MockArtifact{
		BuilderIdValue: "invalid.builder",
	}

	config := testGoodConfig()
	server := newSecureServer("foo", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	var p PostProcessor

	p.Configure(config)
	_, _, _, err := p.PostProcess(context.Background(), testUi(), artifact)
	if !strings.Contains(err.Error(), "Unknown artifact type") {
		t.Fatalf("Should error with message 'Unknown artifact type...' with BuilderId: %s", artifact.BuilderIdValue)
	}
}

func TestPostProcessor_PostProcess_checkArtifactFileIsBox(t *testing.T) {
	artifact := &packer.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant", // good
		FilesValue:     []string{"invalid.boxfile"},        // should have .box extension
	}

	config := testGoodConfig()
	server := newSecureServer("foo", nil)
	defer server.Close()
	config["vagrant_cloud_url"] = server.URL
	var p PostProcessor

	p.Configure(config)
	_, _, _, err := p.PostProcess(context.Background(), testUi(), artifact)
	if !strings.Contains(err.Error(), "Unknown files in artifact") {
		t.Fatalf("Should error with message 'Unknown files in artifact...' with artifact file: %s",
			artifact.FilesValue[0])
	}
}

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func TestProviderFromBuilderName(t *testing.T) {
	if providerFromBuilderName("foobar") != "foobar" {
		t.Fatal("should copy unknown provider")
	}

	if providerFromBuilderName("vmware") != "vmware_desktop" {
		t.Fatal("should convert provider")
	}
}

func TestProviderFromVagrantBox_missing_box(t *testing.T) {
	// Bad: Box does not exist
	boxfile := "i_dont_exist.box"
	_, err := providerFromVagrantBox(boxfile)
	if err == nil {
		t.Fatal("Should have error as box file does not exist")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_empty_box(t *testing.T) {
	// Bad: Empty box file
	boxfile, err := newBoxFile()
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatal("Should have error as box file is empty")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_gzip_only_box(t *testing.T) {
	boxfile, err := newBoxFile()
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(boxfile.Name())

	// Bad: Box is just a plain gzip file
	aw := gzip.NewWriter(boxfile)
	_, err = aw.Write([]byte("foo content"))
	if err != nil {
		t.Fatal("Error zipping test box file")
	}
	aw.Close() // Flush the gzipped contents to file

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as box file is a plain gzip file: %s", err)
	}
	t.Logf("%s", err)
}

func newBoxFile() (boxfile *os.File, err error) {
	boxfile, err = ioutil.TempFile(os.TempDir(), "test*.box")
	if err != nil {
		return boxfile, fmt.Errorf("Error creating test box file: %s", err)
	}
	return boxfile, nil
}
