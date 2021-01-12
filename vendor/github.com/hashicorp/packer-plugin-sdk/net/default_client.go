package net

import (
	"net/http"
)

func HttpClientWithEnvironmentProxy() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	return httpClient
}
