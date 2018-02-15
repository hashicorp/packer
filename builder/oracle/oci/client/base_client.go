package oci

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// baseClient provides a basic (AND INTENTIONALLY INCOMPLETE) JSON REST client
// that abstracts away some of the repetitive code required in the OCI Client.
type baseClient struct {
	httpClient  *http.Client
	method      string
	url         string
	queryStruct interface{}
	header      http.Header
	body        interface{}
}

// newBaseClient constructs a default baseClient.
func newBaseClient() *baseClient {
	return &baseClient{
		httpClient: http.DefaultClient,
		method:     "GET",
		header:     make(http.Header),
	}
}

// New creates a copy of an existing baseClient.
func (c *baseClient) New() *baseClient {
	// Copy headers
	header := make(http.Header)
	for k, v := range c.header {
		header[k] = v
	}

	return &baseClient{
		httpClient: c.httpClient,
		method:     c.method,
		url:        c.url,
		header:     header,
	}
}

// Client sets the http Client used to perform requests.
func (c *baseClient) Client(httpClient *http.Client) *baseClient {
	if httpClient == nil {
		c.httpClient = http.DefaultClient
	} else {
		c.httpClient = httpClient
	}
	return c
}

// Base sets the base client url.
func (c *baseClient) Base(path string) *baseClient {
	c.url = path
	return c
}

// Path extends the client url.
func (c *baseClient) Path(path string) *baseClient {
	baseURL, baseErr := url.Parse(c.url)
	pathURL, pathErr := url.Parse(path)
	// Bail on parsing error leaving the client's url unmodified
	if baseErr != nil || pathErr != nil {
		return c
	}

	c.url = baseURL.ResolveReference(pathURL).String()
	return c
}

// QueryStruct sets the struct from which the request querystring is built.
func (c *baseClient) QueryStruct(params interface{}) *baseClient {
	c.queryStruct = params
	return c
}

// SetBody wraps a given struct for serialisation and sets the client body.
func (c *baseClient) SetBody(params interface{}) *baseClient {
	c.body = params
	return c
}

// Header

// AddHeader adds a HTTP header to the client. Existing keys will be extended.
func (c *baseClient) AddHeader(key, value string) *baseClient {
	c.header.Add(key, value)
	return c
}

// SetHeader sets a HTTP header on the client. Existing keys will be
// overwritten.
func (c *baseClient) SetHeader(key, value string) *baseClient {
	c.header.Add(key, value)
	return c
}

// HTTP methods (subset)

// Get sets the client's HTTP method to GET.
func (c *baseClient) Get(path string) *baseClient {
	c.method = "GET"
	return c.Path(path)
}

// Post sets the client's HTTP method to POST.
func (c *baseClient) Post(path string) *baseClient {
	c.method = "POST"
	return c.Path(path)
}

// Delete sets the client's HTTP method to DELETE.
func (c *baseClient) Delete(path string) *baseClient {
	c.method = "DELETE"
	return c.Path(path)
}

// Do executes a HTTP request and returns the response encoded as either error
// or success values.
func (c *baseClient) Do(req *http.Request, successV, failureV interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if successV != nil {
			err = json.NewDecoder(resp.Body).Decode(successV)
		}
	} else {
		if failureV != nil {
			err = json.NewDecoder(resp.Body).Decode(failureV)
		}
	}

	return resp, err
}

// Request builds a http.Request from the baseClient instance.
func (c *baseClient) Request() (*http.Request, error) {
	reqURL, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}

	if c.queryStruct != nil {
		err = addQueryStruct(reqURL, c.queryStruct)
		if err != nil {
			return nil, err
		}
	}

	body := &bytes.Buffer{}
	if c.body != nil {
		if err := json.NewEncoder(body).Encode(c.body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(c.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	// Add headers to request
	for k, vs := range c.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

// Recieve creates a http request from the client and executes it returning the
// response.
func (c *baseClient) Receive(successV, failureV interface{}) (*http.Response, error) {
	req, err := c.Request()
	if err != nil {
		return nil, err
	}
	return c.Do(req, successV, failureV)
}

// addQueryStruct converts a struct to a querystring and merges any values
// provided in the URL itself.
func addQueryStruct(reqURL *url.URL, queryStruct interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	queryValues, err := query.Values(queryStruct)
	if err != nil {
		return err
	}

	for k, vs := range queryValues {
		for _, v := range vs {
			urlValues.Add(k, v)
		}
	}
	reqURL.RawQuery = urlValues.Encode()
	return nil
}
