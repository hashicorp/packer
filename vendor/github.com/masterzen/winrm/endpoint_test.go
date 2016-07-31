package winrm

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *WinRMSuite) TestEndpointUrlHttp(c *C) {
	endpoint := &Endpoint{Host: "abc", Port: 123}
	c.Assert(endpoint.url(), Equals, "http://abc:123/wsman")
}

func (s *WinRMSuite) TestEndpointUrlHttps(c *C) {
	endpoint := &Endpoint{Host: "abc", Port: 123, HTTPS: true}
	c.Assert(endpoint.url(), Equals, "https://abc:123/wsman")
}

func (s *WinRMSuite) TestEndpointWithDefaultTimeout(c *C) {
	endpoint := NewEndpoint("test", 5585, false, false, nil, nil, nil, 0)
	c.Assert(endpoint.Timeout, Equals, 60*time.Second)
}

func (s *WinRMSuite) TestEndpointWithTimeout(c *C) {
	endpoint := NewEndpoint("test", 5585, false, false, nil, nil, nil, 120*time.Second)
	c.Assert(endpoint.Timeout, Equals, 120*time.Second)
}
