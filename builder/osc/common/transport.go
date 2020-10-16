package common

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// Transport is used to sing the user for each request
type Transport struct {
	transport http.RoundTripper
	signer    *v4.Signer
	region    string
}

func (t *Transport) sign(req *http.Request, body []byte) error {
	reader := strings.NewReader(string(body))
	timestamp := time.Now()
	_, err := t.signer.Sign(req, reader, "osc", t.region, timestamp)
	return err
}

// RoundTrip is implemented according with the interface RoundTrip to sing for each request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	//Get the body
	getBody := req.GetBody
	copyBody, err := getBody()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(copyBody)
	if err != nil {
		return nil, err
	}

	if err := t.sign(req, body); err != nil {
		return nil, err
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// NewTransport returns the transport signing with the given credentials
func NewTransport(accessKey, accessSecret, region string, t http.RoundTripper) *Transport {
	s := &v4.Signer{
		Credentials: credentials.NewStaticCredentials(accessKey,
			accessSecret, ""),
	}
	return &Transport{t, s, region}
}
