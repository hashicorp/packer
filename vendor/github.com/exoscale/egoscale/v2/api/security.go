package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// SecurityProviderExoscale represents an Exoscale public API security
// provider.
type SecurityProviderExoscale struct {
	// ReqExpire represents the request expiration duration.
	ReqExpire time.Duration

	apiKey    string
	apiSecret string
}

// NewSecurityProvider returns a new Exoscale public API security
// provider to sign API requests using the specified API key/secret.
func NewSecurityProvider(apiKey, apiSecret string) (*SecurityProviderExoscale, error) {
	if apiKey == "" {
		return nil, errors.New("missing API key")
	}

	if apiSecret == "" {
		return nil, errors.New("missing API secret")
	}

	return &SecurityProviderExoscale{
		ReqExpire: 10 * time.Minute,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

// Intercept is an HTTP middleware that intercepts and signs client requests
// before sending them to the API endpoint.
func (s *SecurityProviderExoscale) Intercept(_ context.Context, req *http.Request) error {
	return s.signRequest(req, time.Now().UTC().Add(s.ReqExpire))
}

func (s *SecurityProviderExoscale) signRequest(req *http.Request, expiration time.Time) error {
	var (
		sigParts    []string
		headerParts []string
	)

	// Request method/URL path
	sigParts = append(sigParts, fmt.Sprintf("%s %s", req.Method, req.URL.Path))
	headerParts = append(headerParts, "EXO2-HMAC-SHA256 credential="+s.apiKey)

	// Request body if present
	body := ""
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		err = req.Body.Close()
		if err != nil {
			return err
		}
		body = string(data)
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
	}
	sigParts = append(sigParts, body)

	// Request query string parameters
	// Important: this is order-sensitive, we have to have to sort parameters alphabetically to ensure signed
	// values match the names listed in the "signed-query-args=" signature pragma.
	signedParams, paramsValues := extractRequestParameters(req)
	sigParts = append(sigParts, paramsValues)
	if len(signedParams) > 0 {
		headerParts = append(headerParts, "signed-query-args="+strings.Join(signedParams, ";"))
	}

	// Request headers -- none at the moment
	// Note: the same order-sensitive caution for query string parameters applies to headers.
	sigParts = append(sigParts, "")

	// Request expiration date (UNIX timestamp, no line return)
	sigParts = append(sigParts, fmt.Sprint(expiration.Unix()))
	headerParts = append(headerParts, "expires="+fmt.Sprint(expiration.Unix()))

	h := hmac.New(sha256.New, []byte(s.apiSecret))
	if _, err := h.Write([]byte(strings.Join(sigParts, "\n"))); err != nil {
		return err
	}
	headerParts = append(headerParts, "signature="+base64.StdEncoding.EncodeToString(h.Sum(nil)))

	req.Header.Set("Authorization", strings.Join(headerParts, ","))

	return nil
}

// extractRequestParameters returns the list of request URL parameters names
// and a strings concatenating the values of the parameters.
func extractRequestParameters(req *http.Request) ([]string, string) {
	var (
		names  []string
		values string
	)

	for param, values := range req.URL.Query() {
		// Keep only parameters that hold exactly 1 value (i.e. no empty or multi-valued parameters)
		if len(values) == 1 {
			names = append(names, param)
		}
	}
	sort.Strings(names)

	for _, param := range names {
		values += req.URL.Query().Get(param)
	}

	return names, values
}
