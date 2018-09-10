package ssh

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// ConnectFunc is a convenience method for returning a function
// that just uses net.Dial to communicate with the remote end that
// is suitable for use with the SSH communicator configuration.
func ConnectFunc(network, addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		c, err := net.DialTimeout(network, addr, 15*time.Second)
		if err != nil {
			return nil, err
		}

		if tcpConn, ok := c.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(5 * time.Second)
		}

		return c, nil
	}
}

// ProxyConnectFunc is a convenience method for returning a function
// that connects to a host using SOCKS5 proxy
func ProxyConnectFunc(socksProxy string, socksAuth *proxy.Auth, network, addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		// create a socks5 dialer
		dialer, err := proxy.SOCKS5("tcp", socksProxy, socksAuth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("Can't connect to the proxy: %s", err)
		}

		c, err := dialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		return c, nil
	}
}

// ProxyConnectFunc is a convenience method for returning a function
// that connects to a host using an HTTP CONNECT proxy
//
// Note: shamelessly copied from https://github.com/golang/build/blob/ce623a5/cmd/buildlet/reverse.go#L213
func HTTPProxyConnectFunc(proxyURL *url.URL, addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		proxyAddr := proxyURL.Host
		if proxyURL.Port() == "" {
			proxyAddr = net.JoinHostPort(proxyAddr, "80")
		}

		c, err := net.Dial("tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("dialing proxy %q failed: %v", proxyAddr, err)
		}

		req := bytes.NewBuffer(nil)
		fmt.Fprintf(req, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, proxyURL.Hostname())
		if proxyURL.User != nil {
			auth := proxyURL.User.Username()
			if p, ok := proxyURL.User.Password(); ok {
				auth += ":" + p
			}

			fmt.Fprintf(req, "Proxy-Authorization: Basic %s\r\n", base64.StdEncoding.EncodeToString([]byte(auth)))
		}
		fmt.Fprintf(req, "\r\n")

		req.WriteTo(c)

		br := bufio.NewReader(c)
		res, err := http.ReadResponse(br, nil)
		if err != nil {
			return nil, fmt.Errorf("reading HTTP response from CONNECT to %s via proxy %s failed: %v", addr, proxyAddr, err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("proxy error from %s while dialing %s: %v", proxyAddr, addr, res.Status)
		}

		// It's safe to discard the bufio.Reader here and return the
		// original TCP conn directly because we only use this for
		// TLS, and in TLS the client speaks first, so we know there's
		// no unbuffered data. But we can double-check.
		if br.Buffered() > 0 {
			return nil, fmt.Errorf("unexpected %d bytes of buffered data from CONNECT proxy %q", br.Buffered(), proxyAddr)
		}

		return c, nil
	}
}

// BastionConnectFunc is a convenience method for returning a function
// that connects to a host over a bastion connection.
func BastionConnectFunc(
	connFunc func() (net.Conn, error),
	bConf *ssh.ClientConfig,
	proto string,
	addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		sshConn, err := connFunc()
		if err != nil {
			return nil, fmt.Errorf("Error connecting to bastion: %s", err)
		}

		sshClientConn, sshChans, sshReqs, err := ssh.NewClientConn(sshConn, addr, bConf)
		if err != nil {
			return nil, fmt.Errorf("Error initialising bastion client: %s", err)
		}

		bastion := ssh.NewClient(sshClientConn, sshChans, sshReqs)

		// Connect through to the end host
		conn, err := bastion.Dial(proto, addr)
		if err != nil {
			bastion.Close()
			return nil, err
		}

		// Wrap it up so we close both things properly
		return &bastionConn{
			Conn:    conn,
			Bastion: bastion,
		}, nil
	}
}

type bastionConn struct {
	net.Conn
	Bastion *ssh.Client
}

func (c *bastionConn) Close() error {
	c.Conn.Close()
	return c.Bastion.Close()
}
