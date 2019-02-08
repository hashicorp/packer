package common

import (
	"net/http"

	"github.com/hashicorp/go-rootcerts"
)

func HttpClientWithEnvironmentProxy() *http.Client {
	httpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	rootcerts.ConfigureTLS(httpTransport.TLSClientConfig, nil)

	httpClient := &http.Client{
		Transport: httpTransport,
	}

	return httpClient
}
