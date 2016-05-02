// Package http implements a HTTP client for go-git.
package http

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/src-d/go-git.v3/clients/common"

	"gopkg.in/src-d/go-git.v3/core"
)

var InvalidAuthMethodErr = errors.New("invalid http auth method: a http.HTTPAuthMethod should be provided.")

type HTTPAuthMethod interface {
	common.AuthMethod
	setAuth(r *http.Request)
}

type BasicAuth struct {
	username, password string
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{username, password}
}

func (a *BasicAuth) setAuth(r *http.Request) {
	r.SetBasicAuth(a.username, a.password)
}

func (a *BasicAuth) Name() string {
	return "http-basic-auth"
}

func (a *BasicAuth) String() string {
	masked := "*******"
	if a.password == "" {
		masked = "<empty>"
	}

	return fmt.Sprintf("%s - %s:%s", a.Name(), a.username, masked)
}

type HTTPError struct {
	Response *http.Response
}

func NewHTTPError(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	err := &HTTPError{r}
	if r.StatusCode == 404 || r.StatusCode == 401 {
		return core.NewPermanentError(common.NotFoundErr)
	}

	return core.NewUnexpectedError(err)
}

func (e *HTTPError) StatusCode() int {
	return e.Response.StatusCode
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unexpected requesting %q status code: %d",
		e.Response.Request.URL, e.Response.StatusCode,
	)
}
