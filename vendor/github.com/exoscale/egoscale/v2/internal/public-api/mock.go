package publicapi

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	*httpmock.MockTransport
	ClientWithResponsesInterface
}

func NewMockClient() *MockClient {
	var c MockClient

	c.MockTransport = httpmock.NewMockTransport()

	return &c
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	hc := http.Client{Transport: c.MockTransport}

	return hc.Do(req)
}
