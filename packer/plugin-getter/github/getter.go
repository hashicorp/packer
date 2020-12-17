package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

type Getter struct {
	Client *http.Client
}

var _ plugingetter.Getter = &Getter{}

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

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	if g.Client == nil {
		g.Client = &http.Client{}
	}

	mockVersions := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(Versions)
		}),
	)

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
		req, err := http.NewRequest("GET", mockVersions.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := g.Client.Do(req)
		if err != nil {
			return nil, err
		}
		return transformVersionStream(resp.Body)
	case "sha256":
		req, err := http.NewRequest("GET", sha256.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := g.Client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	case "binary":
		req, err := http.NewRequest("GET", binary.URL, nil)
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
