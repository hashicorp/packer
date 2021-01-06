package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v33/github"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"golang.org/x/oauth2"
)

const (
	ghTokenAccessor  = "PKR_GITHUB_API_TOKEN"
	defaultUserAgent = "curl/7.64.1"
	defaultMediaType = "application/octet-stream"
)

type Getter struct {
	Client *github.Client
}

var _ plugingetter.Getter = &Getter{}

// transformVersionStream get a stream from github tags and transforms it into
// something Packer wants, namely a json list of Release.
func transformVersionStream(in io.ReadCloser) (io.ReadCloser, error) {
	if in == nil {
		return nil, fmt.Errorf("transformVersionStream got nil body")
	}
	defer in.Close()
	dec := json.NewDecoder(in)

	m := []struct {
		Ref string `json:"ref"`
	}{}
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	out := []plugingetter.Release{}
	for _, m := range m {
		out = append(out, plugingetter.Release{
			Version: strings.TrimPrefix(m.Ref, "refs/tags/"),
		})
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(out); err != nil {
		return nil, err
	}

	return ioutil.NopCloser(buf), nil
}

type HostSpecificTokenAuthTransport struct {
	TokenSources map[string]oauth2.TokenSource

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper

	Base http.RoundTripper
}

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Source.
func (t *HostSpecificTokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	source, found := t.TokenSources[req.Host]
	if found {
		reqBodyClosed := false
		if req.Body != nil {
			defer func() {
				if !reqBodyClosed {
					req.Body.Close()
				}
			}()
		}

		if source == nil {
			return nil, errors.New("transport's Source is nil")
		}
		token, err := source.Token()
		if err != nil {
			return nil, err
		}

		token.SetAuthHeader(req)

		// req.Body is assumed to be closed by the base RoundTripper.
		reqBodyClosed = true
	}

	return t.base().RoundTrip(req)
}

func (t *HostSpecificTokenAuthTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	ctx := context.TODO()
	log.Printf("[TRACE] github.get %s", what)
	if g.Client == nil {
		var tc *http.Client
		if tk := os.Getenv(ghTokenAccessor); tk != "" {
			log.Printf("[TRACE] Using Github token")
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: tk},
			)
			tc = &http.Client{
				Transport: &HostSpecificTokenAuthTransport{
					TokenSources: map[string]oauth2.TokenSource{
						"api.github.com": ts,
					},
				},
			}
		}
		g.Client = github.NewClient(tc)
		g.Client.UserAgent = defaultUserAgent
	}

	binary := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("1"))
		}),
	)

	var req *http.Request
	var err error
	transform := func(in io.ReadCloser) (io.ReadCloser, error) {
		return in, nil
	}

	switch what {
	case "releases":
		req, err = g.Client.NewRequest("GET", filepath.Join("/repos/", opts.PluginRequirement.Identifier.RealRelativePath(), "/git/matching-refs/tags"), nil)
		req.Header.Set("User-Agent", "Potato")
		transform = transformVersionStream
	case "sha256":
		// something like https://github.com/azr/packer-plugin-amazon/releases/download/v0.0.1/sha256
		req, err = g.Client.NewRequest("GET", "https://github.com/azr/packer-plugin-amazon/releases/download/v0.0.1/sha256", nil)
		header := req.Header
		header.Del("Authorization")
		req.Header = header
	case "binary":
		req, err = g.Client.NewRequest("GET", binary.URL, nil)
		header := req.Header
		header.Del("Authorization")
		req.Header = header
	default:
		return nil, fmt.Errorf("%q not implemented", what)
	}
	if err != nil {
		return nil, err
	}
	resp, err := g.Client.BareDo(ctx, req)
	if err != nil {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[TRACE] Request %#v failed: %#v. Resp: %s", req, err, string(b))
		return nil, err
	}

	if c := resp.StatusCode; 200 <= c && c <= 299 {
		return transform(resp.Body)
	}

	defer resp.Body.Close()
	log.Printf("[TRACE] Request %#v failed: %v", req, resp.Status)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Printf("Request failed: %s", string(b))
	return nil, err
}
