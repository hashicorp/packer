/*
Package http is an implementation of http protocol
*/
package http

import (
	"io/ioutil"
	"net/http"
	"time"
)

// Client is the interface of http client
type Client interface {
	Send(*HttpRequest) (*HttpResponse, error)
}

// HttpClient used to send a real request via http to server
type HttpClient struct {
}

// NewHttpClient will create a new HttpClient instance
func NewHttpClient() HttpClient {
	return HttpClient{}
}

// Send will send a real http request to remote server
func (c *HttpClient) Send(req *HttpRequest) (*HttpResponse, error) {
	// build http.Client with timeout settings
	httpClient, err := c.buildHTTPClient(req.GetTimeout())
	if err != nil {
		return nil, err
	}

	// convert sdk http request to origin http.Request
	httpReq, err := req.buildHTTPRequest()
	if err != nil {
		return nil, err
	}

	// TODO: enable tracer via `httptrace` package
	resp, err := c.doHTTPRequest(httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *HttpClient) buildHTTPClient(timeout time.Duration) (*http.Client, error) {
	httpClient := http.Client{}
	if timeout != 0 {
		httpClient = http.Client{Timeout: timeout}
	}
	return &httpClient, nil
}

func (c *HttpClient) doHTTPRequest(client *http.Client, req *http.Request) (*HttpResponse, error) {
	// send request
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// raise status error
	if httpResp.StatusCode >= 400 {
		return nil, NewStatusError(httpResp.StatusCode, httpResp.Status)
	}

	// read content
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// build response wrapper
	resp := NewHttpResponse()
	resp.setHttpResponse(httpResp)
	resp.SetBody(body)
	return resp, nil
}
