// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

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

// setup sets up a test HTTP server along with a bmcs.Client that is
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
