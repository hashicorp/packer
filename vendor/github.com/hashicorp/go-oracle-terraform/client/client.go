package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/hashicorp/go-oracle-terraform/opc"
)

const DEFAULT_MAX_RETRIES = 1
const USER_AGENT_HEADER = "User-Agent"

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
		client.MaxRetries = opc.Int(DEFAULT_MAX_RETRIES)
	}

	// Protect against any nil http client
	if c.HTTPClient == nil {
		return nil, fmt.Errorf("No HTTP client specified in config")
	}

	return client, nil
}

// Marshalls the request body and returns the resulting byte slice
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

// Builds an HTTP Request that accepts a pre-marshaled body parameter as a raw byte array
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
	req.Header.Add(USER_AGENT_HEADER, *c.UserAgent)

	return req, nil
}

// Build a new HTTP request that doesn't marshall the request body
func (c *Client) BuildNonJSONRequest(method, path string, body io.ReadSeeker) (*http.Request, error) {
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
	req.Header.Add(USER_AGENT_HEADER, *c.UserAgent)

	return req, nil
}

// This method executes the http.Request from the BuildRequest method.
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
		buf.ReadFrom(resp.Body)
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
		retries = DEFAULT_MAX_RETRIES
	} else {
		retries = *c.MaxRetries
	}

	var statusCode int
	var errMessage string

	for i := 0; i < retries; i++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return resp, err
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp, nil
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		errMessage = buf.String()
		statusCode = resp.StatusCode
		c.DebugLogString(fmt.Sprintf("Encountered HTTP (%d) Error: %s", statusCode, errMessage))
		c.DebugLogString(fmt.Sprintf("%d/%d retries left", i+1, retries))
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

// Retry function
func (c *Client) WaitFor(description string, timeout time.Duration, test func() (bool, error)) error {
	tick := time.Tick(1 * time.Second)

	timeoutSeconds := int(timeout.Seconds())

	for i := 0; i < timeoutSeconds; i++ {
		select {
		case <-tick:
			completed, err := test()
			c.DebugLogString(fmt.Sprintf("Waiting for %s (%d/%ds)", description, i, timeoutSeconds))
			if err != nil || completed {
				return err
			}
		}
	}
	return fmt.Errorf("Timeout waiting for %s", description)
}

// Used to determine if the checked resource was found or not.
func WasNotFoundError(e error) bool {
	err, ok := e.(*opc.OracleError)
	if ok {
		return err.StatusCode == 404
	}
	return false
}
