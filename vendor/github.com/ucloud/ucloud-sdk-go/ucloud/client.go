/*
Package ucloud is a package of utilities to setup ucloud sdk and improve using experience
*/
package ucloud

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/private/utils"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// Client 客户端
type Client struct {
	// configurations
	credential *auth.Credential
	config     *Config

	// composited instances
	httpClient http.Client
	logger     log.Logger

	// internal properties
	requestHandlers      []RequestHandler
	httpRequestHandlers  []HttpRequestHandler
	responseHandlers     []ResponseHandler
	httpResponseHandlers []HttpResponseHandler
}

// NewClient will create an client of ucloud sdk
func NewClient(config *Config, credential *auth.Credential) *Client {
	client := Client{
		credential: credential,
		config:     config,
	}

	client.requestHandlers = append(client.requestHandlers, defaultRequestHandlers...)
	client.httpRequestHandlers = append(client.httpRequestHandlers, defaultHttpRequestHandlers...)
	client.responseHandlers = append(client.responseHandlers, defaultResponseHandlers...)
	client.httpResponseHandlers = append(client.httpResponseHandlers, defaultHttpResponseHandlers...)

	client.logger = log.New()
	client.logger.SetLevel(config.LogLevel)

	return &client
}

// SetHttpClient will setup a http client
func (c *Client) SetHttpClient(httpClient http.Client) error {
	c.httpClient = httpClient
	return nil
}

// GetCredential will return the creadential config of client.
func (c *Client) GetCredential() *auth.Credential {
	return c.credential
}

// GetConfig will return the config of client.
func (c *Client) GetConfig() *Config {
	return c.config
}

// SetLogger will set the logger of client
func (c *Client) SetLogger(logger log.Logger) {
	c.logger = logger
}

// GetLogger will set the logger of client
func (c *Client) GetLogger() log.Logger {
	return c.logger
}

// InvokeAction will do an action request from a request struct and set response value into res struct pointer
func (c *Client) InvokeAction(action string, req request.Common, resp response.Common) error {
	return c.InvokeActionWithPatcher(action, req, resp, utils.RetCodePatcher)
}

// InvokeActionWithPatcher will invoke action by patchers
func (c *Client) InvokeActionWithPatcher(action string, req request.Common, resp response.Common, patches ...utils.Patch) error {
	var err error
	req.SetAction(action)
	req.SetRequestTime(time.Now())
	resp.SetRequest(req)

	for _, handler := range c.requestHandlers {
		req, err = handler(c, req)
		if err != nil {
			return uerr.NewClientError(uerr.ErrInvalidRequest, err)
		}
	}

	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return uerr.NewClientError(uerr.ErrInvalidRequest, err)
	}

	for _, handler := range c.httpRequestHandlers {
		httpReq, err = handler(c, httpReq)
		if err != nil {
			return uerr.NewClientError(uerr.ErrInvalidRequest, err)
		}
	}

	if c.httpClient == nil {
		httpClient := http.NewHttpClient()
		c.httpClient = &httpClient
	}

	httpResp, err := c.httpClient.Send(httpReq)

	// use response middleware to handle http response
	// such as convert some http status to error
	for _, handler := range c.httpResponseHandlers {
		httpResp, err = handler(c, httpReq, httpResp, err)
	}

	if err == nil {
		// use patch object to resolve the http response body
		// in general, it will be fix common server error before server bugfix is released.
		body := httpResp.GetBody()

		for _, patch := range patches {
			body = patch.Patch(body)
		}

		err = c.unmarshalHTTPResponse(body, resp)
	}

	// use response middle to build and convert response when response has been created.
	// such as retry, report traceback, print log and etc.
	for _, handler := range c.responseHandlers {
		resp, err = handler(c, req, resp, err)
	}

	return err
}

// AddHttpRequestHandler will append a response handler to client
func (c *Client) AddHttpRequestHandler(h HttpRequestHandler) error {
	c.httpRequestHandlers = append([]HttpRequestHandler{h}, c.httpRequestHandlers...)
	return nil
}

// AddRequestHandler will append a response handler to client
func (c *Client) AddRequestHandler(h RequestHandler) error {
	c.requestHandlers = append([]RequestHandler{h}, c.requestHandlers...)
	return nil
}

// AddHttpResponseHandler will append a http response handler to client
func (c *Client) AddHttpResponseHandler(h HttpResponseHandler) error {
	c.httpResponseHandlers = append(c.httpResponseHandlers, h)
	return nil
}

// AddResponseHandler will append a response handler to client
func (c *Client) AddResponseHandler(h ResponseHandler) error {
	c.responseHandlers = append(c.responseHandlers, h)
	return nil
}
