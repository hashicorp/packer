package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

const defaultTimeout = 60 * time.Second

// ClientOpt represents a function setting Exoscale API client option.
type ClientOpt func(*Client) error

// ClientOptWithAPIEndpoint returns a ClientOpt overriding the default Exoscale API endpoint.
func ClientOptWithAPIEndpoint(v string) ClientOpt {
	return func(c *Client) error {
		endpointURL, err := url.Parse(v)
		if err != nil {
			return fmt.Errorf("failed to parse URL: %s", err)
		}

		endpointURL = endpointURL.ResolveReference(&url.URL{Path: api.Prefix})
		c.apiEndpoint = endpointURL.String()

		return nil
	}
}

// ClientOptWithTimeout returns a ClientOpt overriding the default client timeout.
func ClientOptWithTimeout(v time.Duration) ClientOpt {
	return func(c *Client) error {
		c.timeout = v

		if v <= 0 {
			return errors.New("timeout value must be greater than 0")
		}

		return nil
	}
}

// ClientOptWithHTTPClient returns a ClientOpt overriding the default http.Client.
// Note: the Exoscale API client will chain additional middleware
// (http.RoundTripper) on the HTTP client internally, which can alter the HTTP
// requests and responses. If you don't want any other middleware than the ones
// currently set to your HTTP client, you should duplicate it and pass a copy
// instead.
func ClientOptWithHTTPClient(v *http.Client) ClientOpt {
	return func(c *Client) error {
		c.httpClient = v

		return nil
	}
}

// Client represents an Exoscale API client.
type Client struct {
	apiKey      string
	apiSecret   string
	apiEndpoint string
	timeout     time.Duration
	httpClient  *http.Client

	*papi.ClientWithResponses
}

// NewClient returns a new Exoscale API client, or an error if one couldn't be initialized.
func NewClient(apiKey, apiSecret string, opts ...ClientOpt) (*Client, error) {
	client := Client{
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		apiEndpoint: api.EndpointURL,
		httpClient:  http.DefaultClient,
		timeout:     defaultTimeout,
	}

	if client.apiKey == "" || client.apiSecret == "" {
		return nil, fmt.Errorf("%w: missing or incomplete API credentials", ErrClientConfig)
	}

	for _, opt := range opts {
		if err := opt(&client); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrClientConfig, err)
		}
	}

	apiSecurityProvider, err := api.NewSecurityProvider(client.apiKey, client.apiSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize API security provider: %s", err)
	}

	apiURL, err := url.Parse(client.apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize API client: %s", err)
	}
	apiURL = apiURL.ResolveReference(&url.URL{Path: api.Prefix})

	client.httpClient.Transport = api.NewAPIErrorHandlerMiddleware(client.httpClient.Transport)

	papiOpts := []papi.ClientOption{
		papi.WithHTTPClient(client.httpClient),
		papi.WithRequestEditorFn(
			papi.MultiRequestsEditor(
				apiSecurityProvider.Intercept,
				setEndpointFromContext,
			),
		),
	}

	if client.ClientWithResponses, err = papi.NewClientWithResponses(apiURL.String(), papiOpts...); err != nil {
		return nil, fmt.Errorf("unable to initialize API client: %s", err)
	}

	return &client, nil
}

// setEndpointFromContext is an HTTP client request interceptor that overrides the "Host" header
// with information from a request endpoint optionally set in the context instance. If none is
// found, the request is left untouched.
func setEndpointFromContext(ctx context.Context, req *http.Request) error {
	if v, ok := ctx.Value(api.ReqEndpoint{}).(api.ReqEndpoint); ok {
		req.Host = v.Host()
		req.URL.Host = v.Host()
	}

	return nil
}
