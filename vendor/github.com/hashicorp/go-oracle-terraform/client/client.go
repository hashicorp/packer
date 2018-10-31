package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-oracle-terraform/opc"
)

const defaultMaxRetries = 1
const userAgentHeader = "User-Agent"

var (
	// defaultUserAgent builds a string containing the Go version, system archityecture and OS,
	// and the go-autorest version.
	defaultUserAgent = fmt.Sprintf("Go/%s (%s-%s) go-oracle-terraform/%s",
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
		Version(),
	)
)

// Client represents an authenticated compute client, with compute credentials and an api client.
type Client struct {
	IdentityDomain *string
	UserName       *string
	Password       *string
	APIEndpoint    *url.URL
	httpClient     *http.Client
	MaxRetries     *int
	UserAgent      *string
	logger         opc.Logger
	loglevel       opc.LogLevelType
}

// NewClient returns a new client
func NewClient(c *opc.Config) (*Client, error) {
	// First create a client
	client := &Client{
		IdentityDomain: c.IdentityDomain,
		UserName:       c.Username,
		Password:       c.Password,
		APIEndpoint:    c.APIEndpoint,
		UserAgent:      &defaultUserAgent,
		httpClient:     c.HTTPClient,
		MaxRetries:     c.MaxRetries,
		loglevel:       c.LogLevel,
	}
	if c.UserAgent != nil {
		client.UserAgent = c.UserAgent
	}

	// Setup logger; defaults to stdout
	if c.Logger == nil {
		client.logger = opc.NewDefaultLogger()
	} else {
		client.logger = c.Logger
	}

	// If LogLevel was not set to something different,
	// double check for env var
	if c.LogLevel == 0 {
		client.loglevel = opc.LogLevel()
	}

	// Default max retries if unset
	if c.MaxRetries == nil {
		client.MaxRetries = opc.Int(defaultMaxRetries)
	}

	// Protect against any nil http client
	if c.HTTPClient == nil {
		return nil, fmt.Errorf("No HTTP client specified in config")
	}

	return client, nil
}

// MarshallRequestBody marshalls the request body and returns the resulting byte slice
// This is split out of the BuildRequestBody method so as to allow
// the developer to print a debug string of the request body if they
// should so choose.
func (c *Client) MarshallRequestBody(body interface{}) ([]byte, error) {
	// Verify interface isnt' nil
	if body == nil {
		return nil, nil
	}

	return json.Marshal(body)
}

// BuildRequestBody builds an HTTP Request that accepts a pre-marshaled body parameter as a raw byte array
// Returns the raw HTTP Request and any error occured
func (c *Client) BuildRequestBody(method, path string, body []byte) (*http.Request, error) {
	// Parse URL Path
	urlPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	var requestBody io.ReadSeeker
	if len(body) != 0 {
		requestBody = bytes.NewReader(body)
	}

	// Create Request
	req, err := http.NewRequest(method, c.formatURL(urlPath), requestBody)
	if err != nil {
		return nil, err
	}
	// Adding UserAgent Header
	req.Header.Add(userAgentHeader, *c.UserAgent)

	return req, nil
}

// BuildNonJSONRequest builds a new HTTP request that doesn't marshall the request body
func (c *Client) BuildNonJSONRequest(method, path string, body io.Reader) (*http.Request, error) {
	// Parse URL Path
	urlPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequest(method, c.formatURL(urlPath), body)
	if err != nil {
		return nil, err
	}
	// Adding UserAgentHeader
	req.Header.Add(userAgentHeader, *c.UserAgent)

	return req, nil
}

// BuildMultipartFormRequest builds a new HTTP Request for a multipart form request from specifies attributes
func (c *Client) BuildMultipartFormRequest(method, path string, files map[string][]byte, parameters map[string]interface{}) (*http.Request, error) {

	urlPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	var (
		part io.Writer
	)
	for fileName, fileContents := range files {
		part, err = writer.CreateFormFile(fileName, fmt.Sprintf("%s.json", fileName))
		if err != nil {
			return nil, err
		}

		_, err = part.Write(fileContents)
		if err != nil {
			return nil, err
		}
	}

	// Add additional parameters to the writer
	for key, val := range parameters {
		if val.(string) != "" {
			_ = writer.WriteField(key, val.(string))
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.formatURL(urlPath), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}

// ExecuteRequest executes the http.Request from the BuildRequest method.
// It is split up to add additional authentication that is Oracle API dependent.
func (c *Client) ExecuteRequest(req *http.Request) (*http.Response, error) {
	// Execute request with supplied client
	resp, err := c.retryRequest(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, nil
	}

	oracleErr := &opc.OracleError{
		StatusCode: resp.StatusCode,
	}

	// Even though the returned body will be in json form, it's undocumented what
	// fields are actually returned. Once we get documentation of the actual
	// error fields that are possible to be returned we can have stricter error types.
	if resp.Body != nil {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return resp, nil
		}
		oracleErr.Message = buf.String()
	}

	// Should return the response object regardless of error,
	// some resources need to verify and check status code on errors to
	// determine if an error actually occurs or not.
	return resp, oracleErr
}

// Allow retrying the request until it either returns no error,
// or we exceed the number of max retries
func (c *Client) retryRequest(req *http.Request) (*http.Response, error) {
	// Double check maxRetries is not nil
	var retries int
	if c.MaxRetries == nil {
		retries = defaultMaxRetries
	} else {
		retries = *c.MaxRetries
	}

	var statusCode int
	var errMessage string

	// Cache the body content for retries.
	// This is to allow reuse of the original request for the retries attempts
	// as the act of reading the body (when doing the httpClient.Do()) closes the
	// Reader.
	var body []byte
	if req.Body != nil {
		var err error
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	// Initial sleep time between retries
	sleep := 1 * time.Second

	for i := retries; i > 0; i-- {

		// replace body with new unread Reader before each request
		if len(body) > 0 {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return resp, err
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp, nil
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return resp, err
		}
		errMessage = buf.String()
		statusCode = resp.StatusCode
		c.DebugLogString(fmt.Sprintf("%s %s Encountered HTTP (%d) Error: %s", req.Method, req.URL, statusCode, errMessage))
		if i != 1 {
			c.DebugLogString(fmt.Sprintf("%d of %d retries remaining. Next retry in %ds", i-1, retries, sleep/time.Second))
			time.Sleep(sleep)
			// increase sleep time for next retry (exponential backoff with jitter)
			// up to a maximum of ~60 seconds
			if sleep <= 30*time.Second {
				jitter := time.Duration(rand.Int63n(int64(sleep))) / 2
				sleep = (sleep * 2) + jitter
			}
		}
	}

	oracleErr := &opc.OracleError{
		StatusCode: statusCode,
		Message:    errMessage,
	}

	// We ran out of retries to make, return the error and response
	return nil, oracleErr
}

func (c *Client) formatURL(path *url.URL) string {
	return c.APIEndpoint.ResolveReference(path).String()
}

// WaitFor - Retry function
func (c *Client) WaitFor(description string, pollInterval, timeout time.Duration, test func() (bool, error)) error {

	timeoutSeconds := int(timeout.Seconds())
	pollIntervalSeconds := int(pollInterval.Seconds())

	c.DebugLogString(fmt.Sprintf("Starting Wait For %s, polling every %d for %d seconds ", description, pollIntervalSeconds, timeoutSeconds))

	for i := 0; i < timeoutSeconds; i += pollIntervalSeconds {
		c.DebugLogString(fmt.Sprintf("Waiting %d seconds for %s (%d/%ds)", pollIntervalSeconds, description, i, timeoutSeconds))
		time.Sleep(pollInterval)
		completed, err := test()
		if err != nil || completed {
			return err
		}
	}
	return fmt.Errorf("Timeout after %d seconds waiting for %s", timeoutSeconds, description)
}

// WasNotFoundError Used to determine if the checked resource was found or not.
func WasNotFoundError(e error) bool {
	err, ok := e.(*opc.OracleError)
	if ok {
		if strings.Contains(err.Error(), "No such service exits") {
			return true
		}
		return err.StatusCode == 404
	}
	return false
}
