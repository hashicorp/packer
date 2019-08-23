package vagrantcloud

import (
	"archive/tar"
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
	"github.com/stretchr/testify/assert"
)

type tarFiles []struct {
	Name, Body string
}

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

func TestProviderFromVagrantBox_no_files_in_archive(t *testing.T) {
	// Bad: Box contains no files
	boxfile, err := createBox(tarFiles{})
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as box file has no contents")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_no_metadata(t *testing.T) {
	// Bad: Box contains no metadata/metadata.json file
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as box file does not include metadata.json file")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_metadata_empty(t *testing.T) {
	// Bad: Create a box with an empty metadata.json file
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", ""},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as box files metadata.json file is empty")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_metadata_bad_json(t *testing.T) {
	// Bad: Create a box with bad JSON in the metadata.json file
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", "{provider: badjson}"},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as box files metadata.json file contains badly formatted JSON")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_metadata_no_provider_key(t *testing.T) {
	// Bad: Create a box with no 'provider' key in the metadata.json file
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"cows":"moo"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as provider key/value pair is missing from boxes metadata.json file")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_metadata_provider_value_empty(t *testing.T) {
	// Bad: The boxes metadata.json file 'provider' key has an empty value
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider":""}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	_, err = providerFromVagrantBox(boxfile.Name())
	if err == nil {
		t.Fatalf("Should have error as value associated with 'provider' key in boxes metadata.json file is empty")
	}
	t.Logf("%s", err)
}

func TestProviderFromVagrantBox_metadata_ok(t *testing.T) {
	// Good: The boxes metadata.json file has the 'provider' key/value pair
	expectedProvider := "virtualbox"
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider":"` + expectedProvider + `"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	provider, err := providerFromVagrantBox(boxfile.Name())
	assert.Equal(t, expectedProvider, provider, "Error: Expected provider: '%s'. Got '%s'", expectedProvider, provider)
	t.Logf("Expected provider '%s'. Got provider '%s'", expectedProvider, provider)
}

func TestGetProvider_artifice(t *testing.T) {
	expectedProvider := "virtualbox"
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider":"` + expectedProvider + `"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("Error creating test box: %s", err)
	}
	defer os.Remove(boxfile.Name())

	provider, err := getProvider("", boxfile.Name(), "artifice")
	assert.Equal(t, expectedProvider, provider, "Error: Expected provider: '%s'. Got '%s'", expectedProvider, provider)
	t.Logf("Expected provider '%s'. Got provider '%s'", expectedProvider, provider)
}

func TestGetProvider_other(t *testing.T) {
	expectedProvider := "virtualbox"

	provider, _ := getProvider(expectedProvider, "foo.box", "other")
	assert.Equal(t, expectedProvider, provider, "Error: Expected provider: '%s'. Got '%s'", expectedProvider, provider)
	t.Logf("Expected provider '%s'. Got provider '%s'", expectedProvider, provider)
}

func newBoxFile() (boxfile *os.File, err error) {
	boxfile, err = ioutil.TempFile(os.TempDir(), "test*.box")
	if err != nil {
		return boxfile, fmt.Errorf("Error creating test box file: %s", err)
	}
	return boxfile, nil
}

func createBox(files tarFiles) (boxfile *os.File, err error) {
	boxfile, err = newBoxFile()
	if err != nil {
		return boxfile, err
	}

	// Box files are gzipped tar archives
	aw := gzip.NewWriter(boxfile)
	tw := tar.NewWriter(aw)

	// Add each file to the box
	for _, file := range files {
		// Create and write the tar file header
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0644,
			Size: int64(len(file.Body)),
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			return boxfile, fmt.Errorf("Error writing box tar file header: %s", err)
		}
		// Write the file contents
		_, err = tw.Write([]byte(file.Body))
		if err != nil {
			return boxfile, fmt.Errorf("Error writing box tar file contents: %s", err)
		}
	}
	// Flush and close each writer
	err = tw.Close()
	if err != nil {
		return boxfile, fmt.Errorf("Error flushing tar file contents: %s", err)
	}
	err = aw.Close()
	if err != nil {
		return boxfile, fmt.Errorf("Error flushing gzip file contents: %s", err)
	}

	return boxfile, nil
}
