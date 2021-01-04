package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

type Getter struct {
	Client *http.Client
}

var _ plugingetter.Getter = &Getter{}

// transformVersionStream get a stream from github tags and transforms it into
// something Packer wants, namely a json list of Release.
func transformVersionStream(in io.ReadCloser) (io.ReadCloser, error) {
	defer in.Close()

	dec := json.NewDecoder(in)

	m := []struct {
		Ref string `json:"ref"`
	}{}
	dec.Decode(&m)

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

func getPluginURL(opts plugingetter.GetOptions) string {
	return "https://api." + opts.PluginRequirement.Identifier.Hostname + "/repos/" + opts.PluginRequirement.Identifier.Namespace + "/" + opts.PluginRequirement.Identifier.Type
}

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	if g.Client == nil {
		g.Client = &http.Client{}
	}
	ctx := context.Background()

	sha256 := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b"))
		}),
	)

	binary := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("1"))
		}),
	)

	switch what {
	case "releases":
		req, err := httpNewRequest(ctx, "GET", getPluginURL(opts)+"/git/matching-refs/tags", nil)
		if err != nil {
			return nil, err
		}
		resp, err := g.Client.Do(req)
		if err != nil {
			return nil, err
		}
		return transformVersionStream(resp.Body)
	case "sha256":
		req, err := httpNewRequest(ctx, "GET", sha256.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := g.Client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	case "binary":
		req, err := httpNewRequest(ctx, "GET", binary.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := g.Client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
	return nil, fmt.Errorf("not implemented")
}

func httpNewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if tk := os.Getenv("HOMEBREW_GITHUB_API_TOKEN"); tk != "" {
		req.SetBasicAuth("username", tk)
	}
	return req, nil
}
