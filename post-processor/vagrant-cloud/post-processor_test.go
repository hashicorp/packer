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
	"runtime"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

type stubResponse struct {
	Path       string
	Method     string
	Response   string
	StatusCode int
}

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

func testNoAccessTokenProvidedConfig() map[string]interface{} {
	return map[string]interface{}{
		"box_tag":             "baz",
		"version_description": "bar",
		"version":             "0.5",
	}
}

func newStackServer(stack []stubResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if len(stack) < 1 {
			rw.Header().Add("Error", fmt.Sprintf("Request stack is empty - Method: %s Path: %s", req.Method, req.URL.Path))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		match := stack[0]
		stack = stack[1:]
		if match.Method != "" && req.Method != match.Method {
			rw.Header().Add("Error", fmt.Sprintf("Request %s != %s", match.Method, req.Method))
			http.Error(rw, fmt.Sprintf("Request %s != %s", match.Method, req.Method), http.StatusInternalServerError)
			return
		}
		if match.Path != "" && match.Path != req.URL.Path {
			rw.Header().Add("Error", fmt.Sprintf("Request %s != %s", match.Path, req.URL.Path))
			http.Error(rw, fmt.Sprintf("Request %s != %s", match.Path, req.URL.Path), http.StatusInternalServerError)
			return
		}
		rw.Header().Add("Complete", fmt.Sprintf("Method: %s Path: %s", match.Method, match.Path))
		rw.WriteHeader(match.StatusCode)
		if match.Response != "" {
			_, err := rw.Write([]byte(match.Response))
			if err != nil {
				panic("failed to write response: " + err.Error())
			}
		}
	}))
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

func newNoAuthServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("authorization") != "" {
			http.Error(rw, "Authorization header was provider", http.StatusBadRequest)
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

func TestPostProcessor_Configure_checkAccessTokenIsRequiredByDefault(t *testing.T) {
	var p PostProcessor
	server := newSecureServer("foo", nil)
	defer server.Close()

	config := testNoAccessTokenProvidedConfig()
	config["vagrant_cloud_url"] = server.URL
	if err := p.Configure(config); err == nil {
		t.Fatalf("Expected access token to be required.")
	}
}

func TestPostProcessor_Configure_checkAccessTokenIsNotRequiredForOverridenVagrantCloud(t *testing.T) {
	var p PostProcessor
	server := newNoAuthServer(nil)
	defer server.Close()

	config := testNoAccessTokenProvidedConfig()
	config["vagrant_cloud_url"] = server.URL
	if err := p.Configure(config); err != nil {
		t.Fatalf("Expected blank access token to be allowed and authenticate to pass: %s", err)
	}
}

func TestPostProcessor_PostProcess_checkArtifactType(t *testing.T) {
	artifact := &packersdk.MockArtifact{
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
	artifact := &packersdk.MockArtifact{
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

func TestPostProcessor_PostProcess_uploadsAndReleases(t *testing.T) {
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider": "virtualbox"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(boxfile.Name())

	artifact := &packersdk.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant",
		FilesValue:     []string{boxfile.Name()},
	}

	s := newStackServer([]stubResponse{stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-path"}})
	defer s.Close()

	stack := []stubResponse{
		stubResponse{StatusCode: 200, Method: "GET", Path: "/authenticate"},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64", Response: `{"tag": "hashicorp/precise64"}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/versions", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/version/0.5/providers", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64/version/0.5/provider/id/upload", Response: `{"upload_path": "` + s.URL + `/box-upload-path"}`},
		stubResponse{StatusCode: 200, Method: "PUT", Path: "/box/hashicorp/precise64/version/0.5/release"},
	}

	server := newStackServer(stack)
	defer server.Close()
	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL
	config["no_direct_upload"] = true

	var p PostProcessor

	err = p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	_, _, _, err = p.PostProcess(context.Background(), testUi(), artifact)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_PostProcess_uploadsAndNoRelease(t *testing.T) {
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider": "virtualbox"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(boxfile.Name())

	artifact := &packersdk.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant",
		FilesValue:     []string{boxfile.Name()},
	}

	s := newStackServer([]stubResponse{stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-path"}})
	defer s.Close()

	stack := []stubResponse{
		stubResponse{StatusCode: 200, Method: "GET", Path: "/authenticate"},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64", Response: `{"tag": "hashicorp/precise64"}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/versions", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/version/0.5/providers", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64/version/0.5/provider/id/upload", Response: `{"upload_path": "` + s.URL + `/box-upload-path"}`},
	}

	server := newStackServer(stack)
	defer server.Close()
	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL
	config["no_direct_upload"] = true
	config["no_release"] = true

	var p PostProcessor

	err = p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, _, _, err = p.PostProcess(context.Background(), testUi(), artifact)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_PostProcess_directUpload5GFile(t *testing.T) {
	// Disable test on Windows due to unreliable sparse file creation
	if runtime.GOOS == "windows" {
		return
	}

	// Boxes up to 5GB are supported for direct upload so
	// set the box asset to be 5GB exactly
	fSize := int64(5368709120)
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider": "virtualbox"}`},
	}
	f, err := createBox(files)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(f.Name())

	if err := expandFile(f, fSize); err != nil {
		t.Fatalf("failed to expand box file - %s", err)
	}

	artifact := &packersdk.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant",
		FilesValue:     []string{f.Name()},
	}
	f.Close()

	s := newStackServer(
		[]stubResponse{
			stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-path"},
		},
	)
	defer s.Close()

	stack := []stubResponse{
		stubResponse{StatusCode: 200, Method: "GET", Path: "/authenticate"},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64", Response: `{"tag": "hashicorp/precise64"}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/versions", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/version/0.5/providers", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64/version/0.5/provider/id/upload/direct"},
		stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-complete"},
	}

	server := newStackServer(stack)
	defer server.Close()
	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL
	config["no_release"] = true

	// Set response here so we have API server URL available
	stack[4].Response = `{"upload_path": "` + s.URL + `/box-upload-path", "callback": "` + server.URL + `/box-upload-complete"}`

	var p PostProcessor

	err = p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, _, _, err = p.PostProcess(context.Background(), testUi(), artifact)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_PostProcess_directUploadOver5GFile(t *testing.T) {
	// Disable test on Windows due to unreliable sparse file creation
	if runtime.GOOS == "windows" {
		return
	}

	// Boxes over 5GB are not supported for direct upload so
	// set the box asset to be one byte over 5GB
	fSize := int64(5368709121)
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider": "virtualbox"}`},
	}
	f, err := createBox(files)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(f.Name())

	if err := expandFile(f, fSize); err != nil {
		t.Fatalf("failed to expand box file - %s", err)
	}
	f.Close()

	artifact := &packersdk.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant",
		FilesValue:     []string{f.Name()},
	}

	s := newStackServer(
		[]stubResponse{
			stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-path"},
		},
	)
	defer s.Close()

	stack := []stubResponse{
		stubResponse{StatusCode: 200, Method: "GET", Path: "/authenticate"},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64", Response: `{"tag": "hashicorp/precise64"}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/versions", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/version/0.5/providers", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64/version/0.5/provider/id/upload", Response: `{"upload_path": "` + s.URL + `/box-upload-path"}`},
	}

	server := newStackServer(stack)
	defer server.Close()
	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL
	config["no_release"] = true

	var p PostProcessor

	err = p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, _, _, err = p.PostProcess(context.Background(), testUi(), artifact)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_PostProcess_uploadsDirectAndReleases(t *testing.T) {
	files := tarFiles{
		{"foo.txt", "This is a foo file"},
		{"bar.txt", "This is a bar file"},
		{"metadata.json", `{"provider": "virtualbox"}`},
	}
	boxfile, err := createBox(files)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer os.Remove(boxfile.Name())

	artifact := &packersdk.MockArtifact{
		BuilderIdValue: "mitchellh.post-processor.vagrant",
		FilesValue:     []string{boxfile.Name()},
	}

	s := newStackServer(
		[]stubResponse{
			stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-path"},
		},
	)
	defer s.Close()

	stack := []stubResponse{
		stubResponse{StatusCode: 200, Method: "GET", Path: "/authenticate"},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64", Response: `{"tag": "hashicorp/precise64"}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/versions", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "POST", Path: "/box/hashicorp/precise64/version/0.5/providers", Response: `{}`},
		stubResponse{StatusCode: 200, Method: "GET", Path: "/box/hashicorp/precise64/version/0.5/provider/id/upload/direct"},
		stubResponse{StatusCode: 200, Method: "PUT", Path: "/box-upload-complete"},
		stubResponse{StatusCode: 200, Method: "PUT", Path: "/box/hashicorp/precise64/version/0.5/release"},
	}

	server := newStackServer(stack)
	defer server.Close()
	config := testGoodConfig()
	config["vagrant_cloud_url"] = server.URL

	// Set response here so we have API server URL available
	stack[4].Response = `{"upload_path": "` + s.URL + `/box-upload-path", "callback": "` + server.URL + `/box-upload-complete"}`

	var p PostProcessor

	err = p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, _, _, err = p.PostProcess(context.Background(), testUi(), artifact)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testUi() *packersdk.BasicUi {
	return &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packersdk.PostProcessor = new(PostProcessor)
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
	if err != nil {
		t.Fatalf("error getting provider from vagrant box %s:%v", boxfile.Name(), err)
	}
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
	if err != nil {
		t.Fatalf("error getting provider %s:%v", boxfile.Name(), err)
	}
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

func expandFile(f *os.File, finalSize int64) (err error) {
	s, err := f.Stat()
	if err != nil {
		return
	}
	size := finalSize - s.Size()
	if size < 1 {
		return
	}
	if _, err = f.Seek(size-1, 2); err != nil {
		return
	}
	if _, err = f.Write([]byte{0}); err != nil {
		return
	}
	return nil
}
