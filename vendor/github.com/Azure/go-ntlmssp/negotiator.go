package ntlmssp

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
)

//Negotiator is a http.Roundtripper decorator that automatically
//converts basic authentication to NTLM/Negotiate authentication when appropriate.
type Negotiator struct{ http.RoundTripper }

//RoundTrip sends the request to the server, handling any authentication
//re-sends as needed.
func (l Negotiator) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// Use default round tripper if not provided
	rt := l.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	// If it is not basic auth, just round trip the request as usual
	reqauth := authheader(req.Header.Get("Authorization"))
	if !reqauth.IsBasic() {
		return rt.RoundTrip(req)
	}
	// Save request body
	body := bytes.Buffer{}
	if req.Body != nil {
		_, err = body.ReadFrom(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))
	}
	// first try anonymous, in case the server still finds us
	// authenticated from previous traffic
	req.Header.Del("Authorization")
	res, err = rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusUnauthorized {
		return res, err
	}

	resauth := authheader(res.Header.Get("Www-Authenticate"))
	if !resauth.IsNegotiate() {
		// Unauthorized, Negotiate not requested, let's try with basic auth
		req.Header.Set("Authorization", string(reqauth))
		res.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

		res, err = rt.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusUnauthorized {
			return res, err
		}
		resauth = authheader(res.Header.Get("Www-Authenticate"))
	}

	if resauth.IsNegotiate() {
		// 401 with request:Basic and response:Negotiate
		res.Body.Close()

		// recycle credentials
		u, p, err := reqauth.GetBasicCreds()
		if err != nil {
			return nil, err
		}

		// send negotiate
		negotiateMessage := NewNegotiateMessage()
		req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(negotiateMessage))
		req.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

		res, err = rt.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// receive challenge?
		resauth = authheader(res.Header.Get("Www-Authenticate"))
		challengeMessage, err := resauth.GetData()
		if err != nil {
			return nil, err
		}
		if !resauth.IsNegotiate() || len(challengeMessage) == 0 {
			// Negotiation failed, let client deal with response
			return res, nil
		}
		res.Body.Close()

		// send authenticate
		authenticateMessage, err := ProcessChallenge(challengeMessage, u, p)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(authenticateMessage))
		req.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

		res, err = rt.RoundTrip(req)
	}

	return res, err
}
