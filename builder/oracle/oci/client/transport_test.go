package oci

import (
	"net/http"
	"strings"
	"testing"
)

const testKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAyLnyzmYj0dZuDo2nclIdEyLZrFZLtw5xFldWpCUl5W3SxKDL
iIgwxpSO75Yf++Rzqc5j6S5bpIrdca6AwVXCNmjjxMPO7vLLm4l4IUOZMv5FqKaC
I2plFz9uBkzGscnYnMbsDA430082E07lYpNv1xy8JwpbrIsqIMh4XCKci/Od5sLR
kEicmOpQK42FGRTQjjmQoWtv+9XED+vYTRL0AxQEC/6i/E7yssFXZ+fpHSKWeKTQ
K/1Fc4pZ1zNzJcDXGuweISx/QMLz78TAPH5OBq/EQzHKSpKvfnWFRyBHne8pwuN8
8wzbbD+/7OFjz28jNSERVJvfYe3X1k69IWMNGwIDAQABAoIBAQCZhcdU38AzxSrG
DMfuYymDslsEObiNWQlbig9lWlhCwx26cDVbxrZvm747NvpdgVyJmqbF+UP0dJVs
Voh51qrFTLIwk4bZMXBTFPCBmJ865knG9RuCFOUew8/WF7C82GHJf0eY7OL7xpDY
cbZ2D8gxofOydHSrYoElM88CwSI00xPCbBKEMrBO94oXC8yfp2bmV6bNhVXwFDEM
qda7M6jVAsBrTOzxUF5VdUUU/MLsu2cCk/ap1zer2Bml9Afk1bMeGJ3XDQgol0pS
CLxaGczpSNVMF9+pjA5sFHR5rmcl0b/kC9HsgOJGhLOimtS94O64dSdWifgsjf6+
fhT2SMiRAoGBAOUDwkdzNqQfvS+qrP2ULpB4vt7MZ70rDGmyMDLZ1VWgcm2cDIad
b7MkFG6GCa48mKkFXK9mOPyq8ELoTjZo2p+relEqf49BpaNxr+cp11pX7g0AkzCa
a8LwdOOUW/poqYl2xLuw9Rz6ky6ybzatMvCwpQCwnbAdABIVxz4oQKHpAoGBAOBg
3uYC/ynGdF9gJTfdS5XPYoLcKKRRspBZrvvDHaWyBnanm5KYwDAZPzLQMqcpyPeo
5xgwMmtNlc6lKKyGkhSLNCV+eO3yAx1h/cq7ityvMS7u6o5sq+/bvtEnbUPYbEtk
AhVD7/w5Yyzzi4beiQxDKe0q1mvUAH56aGqJivBjAoGBALmUMTPbDhUzTwg4Y1Rd
ZtpVrj43H31wS+++oEYktTZc/T0LLi9Llr9w5kmlvmR94CtfF/ted6FwF5/wRajb
kQXAXC83pAR/au0mbCeDhWpFRLculxfUmqxuVBozF9G0TGYDY2rA+++OsgQuPebt
tRDL4/nKJQ4Ygf0lvr4EulM5AoGBALoIdyabu3WmfhwJujH8P8wA+ztmUCgVOIio
YwWIe49C8Er2om1ESqxWcmit6CFi6qY0Gw6Z/2OqGxgPJY8NsBZqaBziJF+cdWqq
MWMiZXqdopi4LC9T+KZROn9tQhGrYfaL/5IkFti3t/uwHbH/1f8dvKhQCSGzz4kN
8n7KdTDjAoGAKh8XdIOQlThFK108VT2yp4BGZXEOvWrR19DLbuUzHVrEX+Bd+uFo
Ruk9iKEH7PSnlRnIvWc1y9qN7fUi5OR3LvQNGlXHyUl6CtmH3/b064bBKudC+XTn
VBelcIoGpH7Dn9I6pKUFauJz1TSbQCIjYGNqrjyzLtG+lH/gy5q4xs8=
-----END RSA PRIVATE KEY-----`

type testTarget struct {
	CapturedReq *http.Request
	Response    *http.Response
}

func (target *testTarget) RoundTrip(req *http.Request) (*http.Response, error) {
	target.CapturedReq = req
	return target.Response, nil
}

func testReq(method string, url string, body string, headers map[string]string) *http.Request {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		panic(err.Error())
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req
}

var KnownGoodCases = []struct {
	name            string
	inputRequest    func() *http.Request
	expectedHeaders map[string]string
}{
	{
		name: "Simple GET",
		inputRequest: func() *http.Request {
			return testReq("GET", "https://example.com/", "", map[string]string{
				"date": "Mon, 26 Sep 2016 11:04:22 GMT"})
		},
		expectedHeaders: map[string]string{
			"date":          "Mon, 26 Sep 2016 11:04:22 GMT",
			"content-type":  "application/json",
			"host":          "example.com",
			"accept":        "application/json",
			"Authorization": "Signature headers=\"date host (request-target)\",keyId=\"tenant/testuser/3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55\",algorithm=\"rsa-sha256\",signature=\"UMw/FImQYZ5JBpfYcR9YN72lhupGl5yS+522NS9glLodU9f4oKRqaocpGdSUSRhhSDKxIx01rV547/HemJ6QqEPaJJuDQPXsGthokWMU2DBGyaMAqhLClgCJiRQMwpg4rdL2tETzkM3wy6UN+I52RYoNSdsnat2ZArCkfl8dIl9ydcwD8/+BqB8d2wyaAIS4iagdPKLAC/Mu9OzyUPOXQhYGYsoEdOowOUkHOlob65PFrlHmKJDdjEF3MDcygEpApItf4iUEloP5bjixAbZEVpj3HLQ5uaPx9m+RsLzYMeO0adE0wOv2YNmwZrExGhXh1BpTU33m5amHeUBxSaG+2A==\",version=\"1\"",
		},
	},
	{
		name: "Simple PUT request",
		inputRequest: func() *http.Request {
			return testReq("PUT", "https://example.com/", "Some Content", map[string]string{
				"date": "Mon, 26 Sep 2016 11:04:22 GMT"})
		},
		expectedHeaders: map[string]string{
			"date":             "Mon, 26 Sep 2016 11:04:22 GMT",
			"content-type":     "application/json",
			"content-length":   "12",
			"x-content-sha256": "lQ8fsURxamLtHxnwTYqd3MNYadJ4ZB/U9yQBKzu/fXA=",
			"accept":           "application/json",
			"Authorization":    "Signature headers=\"(request-target) host date content-length content-type x-content-sha256\",keyId=\"tenant/testuser/3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55\",algorithm=\"rsa-sha256\",signature=\"FHyPt4PE2HGH+iftzcViB76pCJ2R9+DdTCo1Ss4mH4KHQJdyQtPsCpe6Dc19zie6cRr6dsenk21yYnncic8OwZhII8DULj2//qLFGmgFi84s7LJqMQ/COiP7O9KtCN+U8MMt4PV7ZDsiGFn3/8EUJ1wxYscxSIB19S1NpuEL062JgGfkqxTkTPd7V3Xh1NlmUUtQrAMR3l56k1iV0zXY9Uw0CjWYjueMP0JUmkO7zycYAVBrx7Q8wkmejlyD7yFrAnObyEsMm9cIL9IcruWFHeCHFxRLslw7AoLxibAm2Dc9EROuvCK2UkUp8AFkE+QyYDMrrSm1NLBMWdnYqdickA==\",version=\"1\"",
		},
	},
	{
		name: "Simple POST request",
		inputRequest: func() *http.Request {
			return testReq("POST", "https://example.com/", "Some Content", map[string]string{
				"date": "Mon, 26 Sep 2016 11:04:22 GMT"})
		},
		expectedHeaders: map[string]string{
			"date":             "Mon, 26 Sep 2016 11:04:22 GMT",
			"content-type":     "application/json",
			"content-length":   "12",
			"x-content-sha256": "lQ8fsURxamLtHxnwTYqd3MNYadJ4ZB/U9yQBKzu/fXA=",
			"accept":           "application/json",
			"Authorization":    "Signature headers=\"(request-target) host date content-length content-type x-content-sha256\",keyId=\"tenant/testuser/3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55\",algorithm=\"rsa-sha256\",signature=\"WzGIoySkjqydwabMTxjVs05UBu0hThAEBzVs7HbYO45o2XpaoqGiNX67mNzs1PeYrGHpJp8+Ysoz66PChWV/1trxuTU92dQ/FgwvcwBRy5dQvdLkjWCZihNunSk4gt9652w6zZg/ybLon0CFbLRnlanDJDX9BgR3ttuTxf30t5qr2A4fnjFF4VjaU/CzE13cNfaWftjSd+xNcla2sbArF3R0+CEEb0xZEPzTyjjjkyvXdaPZwEprVn8IDmdJvLmRP4EniAPxE1EZIhd712M5ondQkR4/WckM44/hlKDeXGFb4y+QnU02i4IWgOWs3dh2tuzS1hp1zfq7qgPbZ4hp0A==\",version=\"1\"",
		},
	},

	{
		name: "Simple DELETE",
		inputRequest: func() *http.Request {
			return testReq("DELETE", "https://example.com/", "Some Content", map[string]string{
				"date": "Mon, 26 Sep 2016 11:04:22 GMT"})
		},
		expectedHeaders: map[string]string{
			"date":          "Mon, 26 Sep 2016 11:04:22 GMT",
			"content-type":  "application/json",
			"accept":        "application/json",
			"Authorization": "Signature headers=\"date host (request-target)\",keyId=\"tenant/testuser/3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55\",algorithm=\"rsa-sha256\",signature=\"Kj4YSpONZG1cibLbNgxIp4VoS5+80fsB2Fh2Ue28+QyXq4wwrJpMP+8jEupz1yTk1SNPYuxsk7lNOgtI6G1Hq0YJJVum74j46sUwRWe+f08tMJ3c9J+rrzLfpIrakQ8PaudLhHU0eK5kuTZme1dCwRWXvZq3r5IqkGot/OGMabKpBygRv9t0i5ry+bTslSjMqafTWLosY9hgIiGrXD+meB5tpyn+gPVYc//Hc/C7uNNgLJIMk5DKVa4U0YnoY3ojafZTXZQQNGRn2NDMcZUX3f3nJlUIfiZRiOCTkbPwx/fWb4MZtYaEsY5OPficbJRvfOBxSG1wjX+8rgO7ijhMAA==\",version=\"1\"",
		},
	},
}

func TestKnownGoodRequests(t *testing.T) {
	pKey, err := ParsePrivateKey([]byte(testKey), []byte{})
	if err != nil {
		t.Fatalf("Failed to parse test key: %s", err.Error())
	}

	config := &Config{
		Key:         pKey,
		User:        "testuser",
		Tenancy:     "tenant",
		Fingerprint: "3c:b6:44:d7:49:1a:ac:bf:de:7d:76:22:a7:f5:df:55",
	}

	expectedResponse := &http.Response{}
	for _, tt := range KnownGoodCases {
		targetBackend := &testTarget{Response: expectedResponse}
		target := NewTransport(targetBackend, config)

		_, err = target.RoundTrip(tt.inputRequest())
		if err != nil {
			t.Fatalf("%s: Failed to handle request %s", tt.name, err.Error())
		}

		sentReq := targetBackend.CapturedReq

		for header, val := range tt.expectedHeaders {
			if sentReq.Header.Get(header) != val {
				t.Fatalf("%s: Header mismatch in responnse,\n\t expecting \"%s\"\n\t got \"%s\"", tt.name, val, sentReq.Header.Get(header))
			}
		}
	}

}
