package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"net"
	"time"
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

//BastionConnectFunc returns a function that connects to an SSH bastion host,
//then connects to the target address from that bastion. The returned net.Conn
//is suitable for use in creating an ssh.ClientConnection. The underlying ssh
//connection will be closed when connection is closed.
func BastionConnectFunc(bAddr string, bConf *ssh.ClientConfig, addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		bClient, err := ssh.Dial("tcp", bAddr, bConf)
		if err != nil {
			return nil, err
		}

		if conn, err := bClient.Dial("tcp", addr); err != nil {
			bClient.Close()
			return nil, err
		} else {
			return &bastionConn{
				Conn:    conn,
				bastion: bClient,
			}, nil
		}
	}
}

//bastionConn wraps the connection to the host as well as the connection to the
//bastion so that the ssh connection can be torn down when the forwarded
//connection is closed
type bastionConn struct {
	net.Conn
	bastion *ssh.Client
}

func (b *bastionConn) Close() error {
	b.Conn.Close()
	return b.bastion.Close()
}
