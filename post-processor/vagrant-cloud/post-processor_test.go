package vagrantcloud

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
