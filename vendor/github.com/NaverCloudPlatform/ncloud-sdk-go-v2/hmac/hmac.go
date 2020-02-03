package hmac

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"net/url"
)

func NewSigner(secretKey string, hashFunc crypto.Hash) *HMACSigner {
	return &HMACSigner{
		secretKey: secretKey,
		hashFunc:  hashFunc,
	}
}

type signer interface {
	Sign(method string, url string, accessKey string, apiKey string, timestamp string) (string, error)
	HashFunc() crypto.Hash
	Debug(enabled bool)
}

type HMACSigner struct {
	secretKey string
	hashFunc  crypto.Hash
	debug     bool
}

func (s *HMACSigner) Debug(enabled bool) {
	s.debug = enabled
}

func (s *HMACSigner) Sign(method string, reqUrl string, accessKey string, timestamp string) (string, error) {
	const space = " "
	const newLine = "\n"

	u, err := url.Parse(reqUrl)
	if err != nil {
		return "", err
	}

	if s.debug {
		fmt.Println("reqUrl: ", reqUrl)
		fmt.Println("accessKey: ", accessKey)
	}

	h := hmac.New(s.HashFunc().New, []byte(s.secretKey))
	h.Write([]byte(method))
	h.Write([]byte(space))
	h.Write([]byte(u.RequestURI()))
	h.Write([]byte(newLine))
	h.Write([]byte(timestamp))
	h.Write([]byte(newLine))
	h.Write([]byte(accessKey))
	rawSignature := h.Sum(nil)

	base64signature := base64.StdEncoding.EncodeToString(rawSignature)
	if s.debug {
		fmt.Println("Base64 signature:", base64signature)
	}
	return base64signature, nil
}

func (s *HMACSigner) HashFunc() crypto.Hash {
	return s.hashFunc
}
