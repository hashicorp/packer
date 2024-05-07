// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package github

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/hashicorp/packer/hcl2template/addrs"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"golang.org/x/oauth2"
)

const (
	ghTokenAccessor  = "PACKER_GITHUB_API_TOKEN"
	defaultUserAgent = "packer-github-plugin-getter"
	defaultHostname  = "github.com"
)

type Getter struct {
	Client    *github.Client
	UserAgent string
}

var _ plugingetter.Getter = &Getter{}

func transformChecksumStream() func(in io.ReadCloser) (io.ReadCloser, error) {
	return func(in io.ReadCloser) (io.ReadCloser, error) {
		defer in.Close()
		rd := bufio.NewReader(in)
		buffer := bytes.NewBufferString("[")
		json := json.NewEncoder(buffer)
		for i := 0; ; i++ {
			line, err := rd.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return nil, fmt.Errorf(
						"Error reading checksum file: %s", err)
				}
				break
			}
			parts := strings.Fields(line)
			switch len(parts) {
			case 2: // nominal case
				checksumString, checksumFilename := parts[0], parts[1]

				if i > 0 {
					_, _ = buffer.WriteString(",")
				}
				if err := json.Encode(struct {
					Checksum string `json:"checksum"`
					Filename string `json:"filename"`
				}{
					Checksum: checksumString,
					Filename: checksumFilename,
				}); err != nil {
					return nil, err
				}
			}
		}
		_, _ = buffer.WriteString("]")
		return io.NopCloser(buffer), nil
	}
}

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

	return io.NopCloser(buf), nil
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

type GithubPlugin struct {
	Hostname  string
	Namespace string
	Type      string
}

func NewGithubPlugin(source *addrs.Plugin) (*GithubPlugin, error) {
	parts := source.Parts()
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid github.com URI %q: a Github-compatible source must be in the github.com/<namespace>/<name> format.", source.String())
	}

	if parts[0] != defaultHostname {
		return nil, fmt.Errorf("%q doesn't appear to be a valid %q source address; check source and try again.", source.String(), defaultHostname)
	}

	return &GithubPlugin{
		Hostname:  parts[0],
		Namespace: parts[1],
		Type:      strings.Replace(parts[2], "packer-plugin-", "", 1),
	}, nil
}

func (gp GithubPlugin) RealRelativePath() string {
	return path.Join(
		gp.Namespace,
		fmt.Sprintf("packer-plugin-%s", gp.Type),
	)
}

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	ghURI, err := NewGithubPlugin(opts.PluginRequirement.Identifier)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	if g.Client == nil {
		var tc *http.Client
		if tk := os.Getenv(ghTokenAccessor); tk != "" {
			log.Printf("[DEBUG] github-getter: using %s", ghTokenAccessor)
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
		} else {
			log.Printf("[WARNING] github-getter: no GitHub token set, if you intend to install plugins often, please set the %s env var", ghTokenAccessor)
		}
		g.Client = github.NewClient(tc)
		g.Client.UserAgent = defaultUserAgent
		if g.UserAgent != "" {
			g.Client.UserAgent = g.UserAgent
		}
	}

	var req *http.Request
	transform := func(in io.ReadCloser) (io.ReadCloser, error) {
		return in, nil
	}

	switch what {
	case "releases":
		u := filepath.ToSlash("/repos/" + ghURI.RealRelativePath() + "/git/matching-refs/tags")
		req, err = g.Client.NewRequest("GET", u, nil)
		transform = transformVersionStream
	case "sha256":
		// something like https://github.com/sylviamoss/packer-plugin-comment/releases/download/v0.2.11/packer-plugin-comment_v0.2.11_x5_SHA256SUMS
		u := filepath.ToSlash("https://github.com/" + ghURI.RealRelativePath() + "/releases/download/" + opts.Version() + "/" + opts.PluginRequirement.FilenamePrefix() + opts.Version() + "_SHA256SUMS")
		req, err = g.Client.NewRequest(
			"GET",
			u,
			nil,
		)
		transform = transformChecksumStream()
	case "zip":
		u := filepath.ToSlash("https://github.com/" + ghURI.RealRelativePath() + "/releases/download/" + opts.Version() + "/" + opts.ExpectedZipFilename())
		req, err = g.Client.NewRequest(
			"GET",
			u,
			nil,
		)

	default:
		return nil, fmt.Errorf("%q not implemented", what)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] github-getter: getting %q", req.URL)
	resp, err := g.Client.BareDo(ctx, req)
	if err != nil {
		// here BareDo will return an err if the request failed or if the status
		// is not considered a valid http status. So we have to close the body
		// if it's not nil.
		if resp != nil {
			resp.Body.Close()
		}
		switch err := err.(type) {
		case *github.RateLimitError:
			return nil, &plugingetter.RateLimitError{
				SetableEnvVar: ghTokenAccessor,
				Err:           err,
				ResetTime:     err.Rate.Reset.Time,
			}
		default:
			log.Printf("[TRACE] failed requesting: %T. %v", err, err)
			return nil, err
		}

	}

	return transform(resp.Body)
}
