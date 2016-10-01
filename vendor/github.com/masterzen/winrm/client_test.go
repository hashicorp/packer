package winrm

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/masterzen/winrm/soap"

	. "gopkg.in/check.v1"
)

func (s *WinRMSuite) TestNewClient(c *C) {
	endpoint := NewEndpoint("localhost", 5985, false, false, nil, nil, nil, 0)
	client, err := NewClient(endpoint, "Administrator", "v3r1S3cre7")

	c.Assert(err, IsNil)
	c.Assert(client.url, Equals, "http://localhost:5985/wsman")
	c.Assert(client.username, Equals, "Administrator")
	c.Assert(client.password, Equals, "v3r1S3cre7")
}

func (s *WinRMSuite) TestClientCreateShell(c *C) {

	endpoint := NewEndpoint("localhost", 5985, false, false, nil, nil, nil, 0)
	client, err := NewClient(endpoint, "Administrator", "v3r1S3cre7")
	c.Assert(err, IsNil)
	client.http = func(client *Client, message *soap.SoapMessage) (string, error) {
		c.Assert(message.String(), Contains, "http://schemas.xmlsoap.org/ws/2004/09/transfer/Create")
		return createShellResponse, nil
	}

	shell, _ := client.CreateShell()
	c.Assert(shell.id, Equals, "67A74734-DD32-4F10-89DE-49A060483810")
}

func (s *WinRMSuite) TestRun(c *C) {
	ts, host, port, err := runWinRMFakeServer(c, "no input")
	c.Assert(err, IsNil)
	defer ts.Close()

	endpoint := NewEndpoint(host, port, false, false, nil, nil, nil, 0)
	client, err := NewClient(endpoint, "Administrator", "v3r1S3cre7")
	c.Assert(err, IsNil)

	var stdout, stderr bytes.Buffer
	code, err := client.Run("ipconfig /all", &stdout, &stderr)
	c.Assert(err, IsNil)
	c.Assert(code, Equals, 123)
	c.Assert(stdout.String(), Equals, "That's all folks!!!")
	c.Assert(stderr.String(), Equals, "This is stderr, I'm pretty sure!")
}

func (s *WinRMSuite) TestRunWithString(c *C) {
	ts, host, port, err := runWinRMFakeServer(c, "this is the input")
	c.Assert(err, IsNil)
	defer ts.Close()
	endpoint := NewEndpoint(host, port, false, false, nil, nil, nil, 0)
	client, err := NewClient(endpoint, "Administrator", "v3r1S3cre7")
	c.Assert(err, IsNil)

	stdout, stderr, code, err := client.RunWithString("ipconfig /all", "this is the input")
	c.Assert(err, IsNil)
	c.Assert(code, Equals, 123)
	c.Assert(stdout, Equals, "That's all folks!!!")
	c.Assert(stderr, Equals, "This is stderr, I'm pretty sure!")
}

func (s *WinRMSuite) TestRunWithInput(c *C) {
	ts, host, port, err := runWinRMFakeServer(c, "this is the input")
	c.Assert(err, IsNil)
	defer ts.Close()

	endpoint := NewEndpoint(host, port, false, false, nil, nil, nil, 0)
	client, err := NewClient(endpoint, "Administrator", "v3r1S3cre7")
	c.Assert(err, IsNil)

	var stdout, stderr bytes.Buffer
	code, err := client.RunWithInput("ipconfig /all", &stdout, &stderr, strings.NewReader("this is the input"))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, 123)
	c.Assert(stdout.String(), Equals, "That's all folks!!!")
	c.Assert(stderr.String(), Equals, "This is stderr, I'm pretty sure!")
}

func (s *WinRMSuite) TestReplaceTransportWithDecorator(c *C) {
	var myrt rtfunc = func(req *http.Request) (*http.Response, error) {
		req.Body.Close()
		header := http.Header{"Content-Type": {"application/soap+xml; charset=UTF-8"}}
		return &http.Response{StatusCode: 500, Header: header, Body: ioutil.NopCloser(strings.NewReader(""))}, nil
	}

	params := NewParameters("PT60S", "en-US", 153600)
	params.TransportDecorator = func(*http.Transport) http.RoundTripper { return myrt }

	endpoint := NewEndpoint("localhost", 5985, false, false, nil, nil, nil, 0)
	client, err := NewClientWithParameters(endpoint, "Administrator", "password", params)
	c.Assert(err, IsNil)
	_, err = client.http(client, soap.NewMessage())
	c.Assert(err.Error(), Contains, "http error: 500")
}

type rtfunc func(*http.Request) (*http.Response, error)

func (f rtfunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
