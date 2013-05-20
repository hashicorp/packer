package ssh

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"net"
	"strings"
	"testing"
)

// private key for mock server
const testServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA19lGVsTqIT5iiNYRgnoY1CwkbETW5cq+Rzk5v/kTlf31XpSU
70HVWkbTERECjaYdXM2gGcbb+sxpq6GtXf1M3kVomycqhxwhPv4Cr6Xp4WT/jkFx
9z+FFzpeodGJWjOH6L2H5uX1Cvr9EDdQp9t9/J32/qBFntY8GwoUI/y/1MSTmMiF
tupdMODN064vd3gyMKTwrlQ8tZM6aYuyOPsutLlUY7M5x5FwMDYvnPDSeyT/Iw0z
s3B+NCyqeeMd2T7YzQFnRATj0M7rM5LoSs7DVqVriOEABssFyLj31PboaoLhOKgc
qoM9khkNzr7FHVvi+DhYM2jD0DwvqZLN6NmnLwIDAQABAoIBAQCGVj+kuSFOV1lT
+IclQYA6bM6uY5mroqcSBNegVxCNhWU03BxlW//BE9tA/+kq53vWylMeN9mpGZea
riEMIh25KFGWXqXlOOioH8bkMsqA8S7sBmc7jljyv+0toQ9vCCtJ+sueNPhxQQxH
D2YvUjfzBQ04I9+wn30BByDJ1QA/FoPsunxIOUCcRBE/7jxuLYcpR+JvEF68yYIh
atXRld4W4in7T65YDR8jK1Uj9XAcNeDYNpT/M6oFLx1aPIlkG86aCWRO19S1jLPT
b1ZAKHHxPMCVkSYW0RqvIgLXQOR62D0Zne6/2wtzJkk5UCjkSQ2z7ZzJpMkWgDgN
ifCULFPBAoGBAPoMZ5q1w+zB+knXUD33n1J+niN6TZHJulpf2w5zsW+m2K6Zn62M
MXndXlVAHtk6p02q9kxHdgov34Uo8VpuNjbS1+abGFTI8NZgFo+bsDxJdItemwC4
KJ7L1iz39hRN/ZylMRLz5uTYRGddCkeIHhiG2h7zohH/MaYzUacXEEy3AoGBANz8
e/msleB+iXC0cXKwds26N4hyMdAFE5qAqJXvV3S2W8JZnmU+sS7vPAWMYPlERPk1
D8Q2eXqdPIkAWBhrx4RxD7rNc5qFNcQWEhCIxC9fccluH1y5g2M+4jpMX2CT8Uv+
3z+NoJ5uDTXZTnLCfoZzgZ4nCZVZ+6iU5U1+YXFJAoGBANLPpIV920n/nJmmquMj
orI1R/QXR9Cy56cMC65agezlGOfTYxk5Cfl5Ve+/2IJCfgzwJyjWUsFx7RviEeGw
64o7JoUom1HX+5xxdHPsyZ96OoTJ5RqtKKoApnhRMamau0fWydH1yeOEJd+TRHhc
XStGfhz8QNa1dVFvENczja1vAoGABGWhsd4VPVpHMc7lUvrf4kgKQtTC2PjA4xoc
QJ96hf/642sVE76jl+N6tkGMzGjnVm4P2j+bOy1VvwQavKGoXqJBRd5Apppv727g
/SM7hBXKFc/zH80xKBBgP/i1DR7kdjakCoeu4ngeGywvu2jTS6mQsqzkK+yWbUxJ
I7mYBsECgYB/KNXlTEpXtz/kwWCHFSYA8U74l7zZbVD8ul0e56JDK+lLcJ0tJffk
gqnBycHj6AhEycjda75cs+0zybZvN4x65KZHOGW/O/7OAWEcZP5TPb3zf9ned3Hl
NsZoFj52ponUM6+99A2CmezFCN16c4mbA//luWF+k3VVqR6BpkrhKw==
-----END RSA PRIVATE KEY-----`

// password implements the ClientPassword interface
type password string

func (p password) Password(user string) (string, error) {
	return string(p), nil
}

var serverConfig = &ssh.ServerConfig{
	PasswordCallback: func(c *ssh.ServerConn, user, pass string) bool {
		return user == "user" && pass == "pass"
	},
}

func init() {
	// Set the private key of the server, required to accept connections
	if err := serverConfig.SetRSAPrivateKey([]byte(testServerPrivateKey)); err != nil {
		panic("unable to set private key: " + err.Error())
	}
}

func newMockLineServer(t *testing.T) string {
	l, err := ssh.Listen("tcp", "127.0.0.1:0", serverConfig)
	if err != nil {
		t.Fatalf("unable to newMockAuthServer: %s", err)
	}
	go func() {
		defer l.Close()
		c, err := l.Accept()
		if err != nil {
			t.Errorf("Unable to accept incoming connection: %v", err)
			return
		}

		if err := c.Handshake(); err != nil {
			// not Errorf because this is expected to
			// fail for some tests.
			t.Logf("Handshaking error: %v", err)
			return
		}

		t.Log("Accepted SSH connection")
		defer c.Close()

		channel, err := c.Accept()
		if err != nil {
			t.Errorf("Unable to accept a channel: %s", err)
			return
		}

		// Just go in a loop now accepting things... we need to
		// do this to handle packets for SSH.
		go func() {
			c.Accept()
		}()

		channel.Accept()
		t.Log("Accepted channel")
		defer channel.Close()

		data := make([]byte, 0)
		_, err = channel.Read(data)
		if err == nil {
			t.Error("should've gotten a request (exec)")
			return
		}

		req, ok := err.(ssh.ChannelRequest)
		if !ok {
			t.Errorf("couldn't convert err to channel request: %s", err)
			return
		}

		if req.Request != "exec" {
			t.Errorf("unexpected request type: %s", req.Request)
			return
		}

		// Ack it
		channel.AckRequest(true)

		// Just respond back with the payload. We trim the first 4 bytes
		// off of here because it is "\x00\x00\x00\t" and I don't really know
		// why.
		payload := strings.TrimSpace(string(req.Payload[4:]))
		response := fmt.Sprintf("ack: %s", payload)
		_, err = channel.Write([]byte(response))
		if err != nil {
			t.Errorf("error writing response: %s", err)
			return
		}
	}()
	return l.Addr().String()
}

func TestNew_Invalid(t *testing.T) {
	clientConfig := &ssh.ClientConfig{
		User: "user",
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(password("i-am-invalid")),
		},
	}

	conn, err := net.Dial("tcp", newMockLineServer(t))
	if err != nil {
		t.Fatalf("unable to dial to remote side: %s", err)
	}

	_, err = New(conn, clientConfig)
	if err == nil {
		t.Fatal("should have had an error connecting")
	}
}

func TestStart(t *testing.T) {
	clientConfig := &ssh.ClientConfig{
		User: "user",
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(password("pass")),
		},
	}

	conn, err := net.Dial("tcp", newMockLineServer(t))
	if err != nil {
		t.Fatalf("unable to dial to remote side: %s", err)
	}

	client, err := New(conn, clientConfig)
	if err != nil {
		t.Fatalf("error connecting to SSH: %s", err)
	}

	remote, err := client.Start("echo foo")
	if err != nil {
		t.Fatalf("error executing command: %s", err)
	}

	// Wait for it to complete
	t.Log("Waiting for command to complete")
	remote.Wait()

	// Should have the correct output
	output := remote.Stdout.(*bytes.Buffer).String()

	if output != "ack: echo foo" {
		t.Fatalf("unknown output: %#v", output)
	}
}
