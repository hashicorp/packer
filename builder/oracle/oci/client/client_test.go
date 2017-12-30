package oci

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"github.com/go-ini/ini"
)

var (
	mux     *http.ServeMux
	client  *Client
	server  *httptest.Server
	keyFile *os.File
)

// setup sets up a test HTTP server along with a oci.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	parsedURL, _ := url.Parse(server.URL)

	config := &Config{}
	config.baseURL = parsedURL.String()

	var cfg *ini.File
	var err error
	cfg, keyFile, err = BaseTestConfig()

	config, err = loadConfigSection(cfg, "DEFAULT", config)
	if err != nil {
		panic(err)
	}

	client, err = NewClient(config)
	if err != nil {
		panic("Failed to instantiate test client")
	}
}

// teardown closes the test HTTP server
func teardown() {
	server.Close()
	os.Remove(keyFile.Name())
}
