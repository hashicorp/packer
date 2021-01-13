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
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v33/github"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"golang.org/x/oauth2"
)

const (
	ghTokenAccessor  = "PKR_GITHUB_API_TOKEN"
	defaultUserAgent = "packer-plugin-getter"
)

type Getter struct {
	Client    *github.Client
	UserAgent string
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

// HostSpecificTokenAuthTransport makes sure the http roundtripper only sets an
// auth token for requests aimed at a specific host.
//
// This helps for example to get release files from Github as Github will
// redirect to s3 which will error if we give it a Github auth token.
type HostSpecificTokenAuthTransport struct {
	// Host to TokenSource map
	TokenSources map[string]oauth2.TokenSource

	// actual RoundTripper, nil means we use the default one from http.
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
		if g.UserAgent != "" {
			g.Client.UserAgent = g.UserAgent
		}
	}

	var req *http.Request
	var err error
	transform := func(in io.ReadCloser) (io.ReadCloser, error) {
		return in, nil
	}

	switch what {
	case "releases":
		req, err = g.Client.NewRequest("GET", filepath.Join("/repos/", opts.PluginRequirement.Identifier.RealRelativePath(), "/git/matching-refs/tags"), nil)
		transform = transformVersionStream
	case "sha256":
		// something like https://github.com/azr/packer-plugin-amazon/releases/download/v0.0.1/packer-plugin-amazon_darwin-amd64_v0.0.1_x5_SHA256SUM
		req, err = g.Client.NewRequest(
			"GET",
			"https://github.com"+opts.PluginRequirement.Identifier.RealRelativePath()+"/releases/download/"+opts.Version+"/"+opts.ExpectedFilename()+"_SHA256SUM",
			nil,
		)
	case "binary":
		req, err = g.Client.NewRequest(
			"GET",
			"https://github.com"+opts.PluginRequirement.Identifier.RealRelativePath()+"/releases/download/"+opts.Version+"/"+opts.ExpectedFilename(),
			nil,
		)

	default:
		return nil, fmt.Errorf("%q not implemented", what)
	}
	if err != nil {
		return nil, err
	}
	resp, err := g.Client.BareDo(ctx, req)
	if err != nil {
		// here BareDo will return an err if the request failed or if the
		// status is not considered a valid http status.
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("[TRACE] Failed to request: %s.", err)
		return nil, err
	}

	return transform(resp.Body)
}
