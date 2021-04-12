package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"runtime"
	"time"

	v2 "github.com/exoscale/egoscale/v2"
)

const (
	// DefaultTimeout represents the default API client HTTP request timeout.
	DefaultTimeout = 60 * time.Second
)

// UserAgent is the "User-Agent" HTTP request header added to outgoing HTTP requests.
var UserAgent = fmt.Sprintf("egoscale/%s (%s; %s/%s)",
	Version,
	runtime.Version(),
	runtime.GOOS,
	runtime.GOARCH)

// Taggable represents a resource to which tags can be attached
//
// This is a helper to fill the resourcetype of a CreateTags call
type Taggable interface {
	// ResourceType is the name of the Taggable type
	ResourceType() string
}

// Deletable represents an Interface that can be "Delete" by the client
type Deletable interface {
	// Delete removes the given resource(s) or throws
	Delete(context context.Context, client *Client) error
}

// Listable represents an Interface that can be "List" by the client
type Listable interface {
	// ListRequest builds the list command
	ListRequest() (ListCommand, error)
}

// Client represents the API client
type Client struct {
	// HTTPClient holds the HTTP client
	HTTPClient *http.Client
	// Endpoint is the HTTP URL
	Endpoint string
	// APIKey is the API identifier
	APIKey string
	// apisecret is the API secret, hence non exposed
	apiSecret string
	// PageSize represents the default size for a paginated result
	PageSize int
	// Timeout represents the default timeout for the async requests
	Timeout time.Duration
	// Expiration representation how long a signed payload may be used
	Expiration time.Duration
	// RetryStrategy represents the waiting strategy for polling the async requests
	RetryStrategy RetryStrategyFunc
	// Logger contains any log, plug your own
	Logger *log.Logger

	// noV2 represents a flag disabling v2.Client embedding.
	noV2 bool

	// Public API secondary client
	*v2.Client
}

// RetryStrategyFunc represents a how much time to wait between two calls to the API
type RetryStrategyFunc func(int64) time.Duration

// IterateItemFunc represents the callback to iterate a list of results, if false stops
type IterateItemFunc func(interface{}, error) bool

// WaitAsyncJobResultFunc represents the callback to wait a results of an async request, if false stops
type WaitAsyncJobResultFunc func(*AsyncJobResult, error) bool

// ClientOpt represents a new Client option.
type ClientOpt func(*Client)

// WithHTTPClient overrides the Client's default HTTP client.
func WithHTTPClient(hc *http.Client) ClientOpt {
	return func(c *Client) { c.HTTPClient = hc }
}

// WithTimeout overrides the Client's default timeout value (DefaultTimeout).
func WithTimeout(d time.Duration) ClientOpt {
	return func(c *Client) { c.Timeout = d }
}

// WithTrace enables the Client's HTTP request tracing.
func WithTrace() ClientOpt {
	return func(c *Client) { c.TraceOn() }
}

// WithoutV2Client disables implicit v2.Client embedding.
func WithoutV2Client() ClientOpt {
	return func(c *Client) { c.noV2 = true }
}

// NewClient creates an Exoscale API client.
// Note: unless the WithoutV2Client() ClientOpt is passed, this function
// initializes a v2.Client embedded into the returned *Client struct
// inheriting the Exoscale API credentials, endpoint and timeout value, but
// not the custom http.Client. The 2 clients must not share the same
// *http.Client, as it can cause middleware clashes.
func NewClient(endpoint, apiKey, apiSecret string, opts ...ClientOpt) *Client {
	client := &Client{
		HTTPClient: &http.Client{
			Transport: &defaultTransport{next: http.DefaultTransport},
		},
		Endpoint:      endpoint,
		APIKey:        apiKey,
		apiSecret:     apiSecret,
		PageSize:      50,
		Timeout:       DefaultTimeout,
		Expiration:    10 * time.Minute,
		RetryStrategy: MonotonicRetryStrategyFunc(2),
		Logger:        log.New(ioutil.Discard, "", 0),
	}

	for _, opt := range opts {
		opt(client)
	}

	if prefix, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		client.Logger = log.New(os.Stderr, prefix, log.LstdFlags)
		client.TraceOn()
	}

	if !client.noV2 {
		v2Client, err := v2.NewClient(
			client.APIKey,
			client.apiSecret,
			v2.ClientOptWithAPIEndpoint(client.Endpoint),
			v2.ClientOptWithTimeout(client.Timeout),

			// Don't use v2.ClientOptWithHTTPClient() with the root API client's http.Client, as the
			// v2.Client uses HTTP middleware that can break callers that expect CS-compatible error
			// responses.
		)
		if err != nil {
			panic(fmt.Sprintf("unable to initialize API V2 client: %s", err))
		}
		client.Client = v2Client
	}

	return client
}

// Do implemements the v2.HttpRequestDoer interface in order to intercept HTTP response before the
// generated code closes its body, giving us a chance to return meaningful error messages from the API.
// This is only relevant for API v2 operations.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// If the request returned a Go error don't bother analyzing the response
		// body, as there probably won't be any (e.g. connection timeout/refused).
		return resp, err
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		var res struct {
			Message string `json:"message"`
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		if json.Valid(data) {
			if err = json.Unmarshal(data, &res); err != nil {
				return nil, fmt.Errorf("error unmarshaling response: %s", err)
			}
		} else {
			res.Message = string(data)
		}

		switch {
		case resp.StatusCode == http.StatusNotFound:
			return nil, ErrNotFound

		case resp.StatusCode >= 400 && resp.StatusCode < 500:
			return nil, fmt.Errorf("%w: %s", ErrInvalidRequest, res.Message)

		case resp.StatusCode >= 500:
			return nil, fmt.Errorf("%w: %s", ErrAPIError, res.Message)
		}
	}

	return resp, nil
}

// Get populates the given resource or fails
func (c *Client) Get(ls Listable) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return c.GetWithContext(ctx, ls)
}

// GetWithContext populates the given resource or fails
func (c *Client) GetWithContext(ctx context.Context, ls Listable) (interface{}, error) {
	gs, err := c.ListWithContext(ctx, ls)
	if err != nil {
		return nil, err
	}

	switch len(gs) {
	case 0:
		return nil, ErrNotFound

	case 1:
		return gs[0], nil

	default:
		return nil, ErrTooManyFound
	}
}

// Delete removes the given resource of fails
func (c *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return c.DeleteWithContext(ctx, g)
}

// DeleteWithContext removes the given resource of fails
func (c *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, c)
}

// List lists the given resource (and paginate till the end)
func (c *Client) List(g Listable) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return c.ListWithContext(ctx, g)
}

// ListWithContext lists the given resources (and paginate till the end)
func (c *Client) ListWithContext(ctx context.Context, g Listable) (s []interface{}, err error) {
	s = make([]interface{}, 0)

	defer func() {
		if e := recover(); e != nil {
			if g == nil || reflect.ValueOf(g).IsNil() {
				err = fmt.Errorf("g Listable shouldn't be nil, got %#v", g)
				return
			}

			panic(e)
		}
	}()

	req, e := g.ListRequest()
	if e != nil {
		err = e
		return
	}
	c.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
		if item != nil {
			s = append(s, item)
			return true
		}
		err = e
		return false
	})

	return
}

func (c *Client) AsyncListWithContext(ctx context.Context, g Listable) (<-chan interface{}, <-chan error) {
	outChan := make(chan interface{}, c.PageSize)
	errChan := make(chan error)

	go func() {
		defer close(outChan)
		defer close(errChan)

		req, err := g.ListRequest()
		if err != nil {
			errChan <- err
			return
		}
		c.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
			if item != nil {
				outChan <- item
				return true
			}
			errChan <- e
			return false
		})
	}()

	return outChan, errChan
}

// Paginate runs the ListCommand and paginates
func (c *Client) Paginate(g Listable, callback IterateItemFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	c.PaginateWithContext(ctx, g, callback)
}

// PaginateWithContext runs the ListCommand as long as the ctx is valid
func (c *Client) PaginateWithContext(ctx context.Context, g Listable, callback IterateItemFunc) {
	req, err := g.ListRequest()
	if err != nil {
		callback(nil, err)
		return
	}

	pageSize := c.PageSize

	page := 1

	for {
		req.SetPage(page)
		req.SetPageSize(pageSize)
		resp, err := c.RequestWithContext(ctx, req)
		if err != nil {
			// in case of 431, the response is knowingly empty
			if errResponse, ok := err.(*ErrorResponse); ok && page == 1 && errResponse.ErrorCode == ParamError {
				break
			}

			callback(nil, err)
			break
		}

		size := 0
		didErr := false
		req.Each(resp, func(element interface{}, err error) bool {
			// If the context was cancelled, kill it in flight
			if e := ctx.Err(); e != nil {
				element = nil
				err = e
			}

			if callback(element, err) {
				size++
				return true
			}

			didErr = true
			return false
		})

		if size < pageSize || didErr {
			break
		}

		page++
	}
}

// APIName returns the name of the given command
func (c *Client) APIName(command Command) string {
	// This is due to a limitation of Go<=1.7
	_, ok := command.(*AuthorizeSecurityGroupEgress)
	_, okPtr := command.(AuthorizeSecurityGroupEgress)
	if ok || okPtr {
		return "authorizeSecurityGroupEgress"
	}

	info, err := info(command)
	if err != nil {
		panic(err)
	}
	return info.Name
}

// APIDescription returns the description of the given command
func (c *Client) APIDescription(command Command) string {
	info, err := info(command)
	if err != nil {
		return "*missing description*"
	}
	return info.Description
}

// Response returns the response structure of the given command
func (c *Client) Response(command Command) interface{} {
	switch c := command.(type) {
	case AsyncCommand:
		return c.AsyncResponse()
	default:
		return command.Response()
	}
}

// TraceOn activates the HTTP tracer
func (c *Client) TraceOn() {
	if _, ok := c.HTTPClient.Transport.(*traceTransport); !ok {
		c.HTTPClient.Transport = &traceTransport{
			next:   c.HTTPClient.Transport,
			logger: c.Logger,
		}
	}
}

// TraceOff deactivates the HTTP tracer
func (c *Client) TraceOff() {
	if rt, ok := c.HTTPClient.Transport.(*traceTransport); ok {
		c.HTTPClient.Transport = rt.next
	}
}

// defaultTransport is the default HTTP client transport.
type defaultTransport struct {
	next http.RoundTripper
}

// RoundTrip executes a single HTTP transaction while augmenting requests with custom headers.
func (t *defaultTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", UserAgent)

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// traceTransport is a client HTTP middleware that dumps HTTP requests and responses content to a logger.
type traceTransport struct {
	logger *log.Logger
	next   http.RoundTripper
}

// RoundTrip executes a single HTTP transaction
func (t *traceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", UserAgent)

	if dump, err := httputil.DumpRequest(req, true); err == nil {
		t.logger.Printf("%s", dump)
	}

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if dump, err := httputil.DumpResponse(resp, true); err == nil {
		t.logger.Printf("%s", dump)
	}

	return resp, nil
}

// MonotonicRetryStrategyFunc returns a function that waits for n seconds for each iteration
func MonotonicRetryStrategyFunc(seconds int) RetryStrategyFunc {
	return func(iteration int64) time.Duration {
		return time.Duration(seconds) * time.Second
	}
}

// FibonacciRetryStrategy waits for an increasing amount of time following the Fibonacci sequence
func FibonacciRetryStrategy(iteration int64) time.Duration {
	var a, b, i, tmp int64
	a = 0
	b = 1
	for i = 0; i < iteration; i++ {
		tmp = a + b
		a = b
		b = tmp
	}
	return time.Duration(a) * time.Second
}
