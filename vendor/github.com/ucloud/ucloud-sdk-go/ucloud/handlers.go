package ucloud

import (
	"math/rand"
	"time"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// RequestHandler receive request and write data into this request memory area
type RequestHandler func(c *Client, req request.Common) (request.Common, error)

// HttpRequestHandler receive http request and return a new http request
type HttpRequestHandler func(c *Client, req *http.HttpRequest) (*http.HttpRequest, error)

// ResponseHandler receive response and write data into this response memory area
type ResponseHandler func(c *Client, req request.Common, resp response.Common, err error) (response.Common, error)

// HttpResponseHandler receive http response and return a new http response
type HttpResponseHandler func(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error)

var defaultRequestHandlers = []RequestHandler{}
var defaultHttpRequestHandlers = []HttpRequestHandler{}
var defaultResponseHandlers = []ResponseHandler{errorHandler, logHandler, retryHandler}
var defaultHttpResponseHandlers = []HttpResponseHandler{errorHTTPHandler, logDebugHTTPHandler}

func retryHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	if !req.GetRetryable() || err == nil || !err.(uerr.Error).Retryable() {
		return resp, err
	}

	retryCount := req.GetRetryCount()

	// if max retries number is reached, stop and raise last error
	if retryCount >= req.GetMaxretries() {
		return resp, err
	}

	req.SetRetryCount(retryCount + 1)

	// use exponential backoff constant as retry delay
	delay := getExpBackoffDelay(retryCount)
	<-time.After(delay)

	// the resp will be changed after invoke
	err = c.InvokeAction(req.GetAction(), req, resp)

	return resp, err
}

func getExpBackoffDelay(retryCount int) time.Duration {
	minTime := 100
	if retryCount > 7 {
		retryCount = 7
	}

	delay := (1 << (uint(retryCount) * 2)) * (rand.Intn(minTime) + minTime)
	return time.Duration(delay) * time.Millisecond
}

// errorHandler will normalize error to several specific error
func errorHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	if err != nil {
		if _, ok := err.(uerr.Error); ok {
			return resp, err
		}
		if uerr.IsNetworkError(err) {
			return resp, uerr.NewClientError(uerr.ErrNetwork, err)
		}
		return resp, uerr.NewClientError(uerr.ErrSendRequest, err)
	}

	if resp.GetRetCode() != 0 {
		return resp, uerr.NewServerCodeError(resp.GetRetCode(), resp.GetMessage())
	}

	return resp, err
}

func errorHTTPHandler(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
	if err == nil {
		return resp, err
	}

	if statusErr, ok := err.(http.StatusError); ok {
		return resp, uerr.NewServerStatusError(statusErr.StatusCode, statusErr.Message)
	}

	return resp, err
}

func logHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	action := req.GetAction()

	// get strictest logging level
	level := c.config.GetActionLevel(action)

	if err != nil && level >= log.WarnLevel {
		c.logger.Warnf("do %s failed, %s", action, err)
	} else if level >= log.InfoLevel {
		c.logger.Infof("do %s successful!", action)
	}
	return resp, err
}

func logDebugHTTPHandler(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
	action := req.GetQuery("Action")

	// logging request
	c.logActionDebugf(action, "%s", req)

	// logging error
	if err != nil {
		c.logActionErrorf(action, "%s", err)
	}

	// logging response code text
	if resp != nil && resp.GetStatusCode() >= 400 {
		c.logActionWarnf(action, "%s", resp.GetStatusCode())
	}

	// logging response body
	if resp != nil && resp.GetStatusCode() < 400 {
		c.logActionDebugf(action, "%s - %v", resp.GetBody(), resp.GetStatusCode())
	}

	return resp, err
}

func (c *Client) logActionErrorf(action, format string, args ...interface{}) {
	level := c.config.GetActionLevel(action)

	if log.ErrorLevel <= level {
		c.logger.Errorf(format, args...)
	}
}

func (c *Client) logActionWarnf(action, format string, args ...interface{}) {
	level := c.config.GetActionLevel(action)

	if log.WarnLevel <= level {
		c.logger.Warnf(format, args...)
	}
}

func (c *Client) logActionDebugf(action, format string, args ...interface{}) {
	level := c.config.GetActionLevel(action)

	if log.DebugLevel <= level {
		c.logger.Debugf(format, args...)
	}
}
