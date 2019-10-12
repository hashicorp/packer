/*
Package auth is the credential utilities of sdk
*/
package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/url"
	"sort"
)

// Credential is the information of credential keys
type Credential struct {
	PublicKey  string
	PrivateKey string
}

// NewCredential will return credential config with default values
func NewCredential() Credential {
	return Credential{}
}

// CreateSign will encode query string to credential signature.
func (c *Credential) CreateSign(query string) string {
	urlValues, err := url.ParseQuery(query)
	if err != nil {
		return ""
	}
	urlValues.Set("PublicKey", c.PublicKey)
	return c.verifyAc(urlValues)
}

// BuildCredentialedQuery will build query string with signature query param.
func (c *Credential) BuildCredentialedQuery(params map[string]string) string {
	urlValues := url.Values{}
	for k, v := range params {
		urlValues.Set(k, v)
	}
	urlValues.Set("PublicKey", c.PublicKey)
	urlValues.Set("Signature", c.verifyAc(urlValues))
	return urlValues.Encode()
}

func (c *Credential) verifyAc(urlValues url.Values) string {
	// sort keys
	var keys []string
	for k := range urlValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	signQuery := ""
	for _, k := range keys {
		signQuery += k + urlValues.Get(k)
	}
	signQuery += c.PrivateKey
	return encodeSha1(signQuery)
}

func encodeSha1(s string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, s)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}
