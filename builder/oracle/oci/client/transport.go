package oci

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

// Transport adds OCI signature authentication to each outgoing request.
type Transport struct {
	transport http.RoundTripper
	config    *Config
}

// NewTransport creates a new Transport to add OCI signature authentication
// to each outgoing request.
func NewTransport(transport http.RoundTripper, config *Config) *Transport {
	return &Transport{
		transport: transport,
		config:    config,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var buf *bytes.Buffer

	if req.Body != nil {
		buf = new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		req.Body = nopCloser{buf}
	}
	if req.Header.Get("date") == "" {
		req.Header.Set("date", time.Now().UTC().Format(http.TimeFormat))
	}
	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}
	if req.Header.Get("accept") == "" {
		req.Header.Set("accept", "application/json")
	}
	if req.Header.Get("host") == "" {
		req.Header.Set("host", req.URL.Host)
	}

	var signheaders []string
	if (req.Method == "PUT" || req.Method == "POST") && buf != nil {
		signheaders = []string{"(request-target)", "host", "date",
			"content-length", "content-type", "x-content-sha256"}

		if req.Header.Get("content-length") == "" {
			req.Header.Set("content-length", strconv.Itoa(buf.Len()))
		}

		hasher := sha256.New()
		hasher.Write(buf.Bytes())
		hash := hasher.Sum(nil)
		req.Header.Set("x-content-sha256", base64.StdEncoding.EncodeToString(hash))
	} else {
		signheaders = []string{"date", "host", "(request-target)"}
	}

	var signbuffer bytes.Buffer
	for idx, header := range signheaders {
		signbuffer.WriteString(header)
		signbuffer.WriteString(": ")

		if header == "(request-target)" {
			signbuffer.WriteString(strings.ToLower(req.Method))
			signbuffer.WriteString(" ")
			signbuffer.WriteString(req.URL.RequestURI())
		} else {
			signbuffer.WriteString(req.Header.Get(header))
		}

		if idx < len(signheaders)-1 {
			signbuffer.WriteString("\n")
		}
	}

	h := sha256.New()
	h.Write(signbuffer.Bytes())
	digest := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, t.config.Key, crypto.SHA256, digest)
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("Signature headers=\"%s\","+
		"keyId=\"%s/%s/%s\","+
		"algorithm=\"rsa-sha256\","+
		"signature=\"%s\","+
		"version=\"1\"",
		strings.Join(signheaders, " "),
		t.config.Tenancy, t.config.User, t.config.Fingerprint,
		base64.StdEncoding.EncodeToString(signature))
	req.Header.Add("Authorization", authHeader)

	return t.transport.RoundTrip(req)
}
