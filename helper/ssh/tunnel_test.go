package ssh

import (
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/sdk-internals/communicator/ssh"
)

const (
	tunnel8080ToLocal     = "8080:localhost:1234"
	tunnel8080ToRemote    = "8080:example.com:80"
	bindRemoteAddress_NYI = "redis:6379:localhost:6379"
)

func TestTCPToLocalTCP(t *testing.T) {
	tun, err := ParseTunnelArgument(tunnel8080ToLocal, ssh.UnsetTunnel)
	if err != nil {
		t.Fatal(err.Error())
	}
	expectedTun := ssh.TunnelSpec{
		Direction:   ssh.UnsetTunnel,
		ForwardAddr: "localhost:1234",
		ForwardType: "tcp",
		ListenAddr:  "localhost:8080",
		ListenType:  "tcp",
	}
	if tun != expectedTun {
		t.Errorf("Parsed tunnel (%v), want %v", tun, expectedTun)
	}
}

func TestTCPToRemoteTCP(t *testing.T) {
	tun, err := ParseTunnelArgument(tunnel8080ToRemote, ssh.UnsetTunnel)
	if err != nil {
		t.Fatal(err.Error())
	}
	expectedTun := ssh.TunnelSpec{
		Direction:   ssh.UnsetTunnel,
		ForwardAddr: "example.com:80",
		ForwardType: "tcp",
		ListenAddr:  "localhost:8080",
		ListenType:  "tcp",
	}
	if tun != expectedTun {
		t.Errorf("Parsed tunnel (%v), want %v", tun, expectedTun)
	}
}

func TestBindAddress_NYI(t *testing.T) {
	tun, err := ParseTunnelArgument(bindRemoteAddress_NYI, ssh.UnsetTunnel)
	if err == nil {
		t.Fatal(err.Error())
	}
	expectedTun := ssh.TunnelSpec{
		Direction:   ssh.UnsetTunnel,
		ForwardAddr: "redis:6379",
		ForwardType: "tcp",
		ListenAddr:  "localhost:6379",
		ListenType:  "tcp",
	}
	if tun == expectedTun {
		t.Errorf("Parsed tunnel (%v), want %v", tun, expectedTun)
	}
}

func TestInvalidTunnels(t *testing.T) {
	invalids := []string{
		"nope:8080",                       // insufficient parts
		"nope:localhost:8080",             // listen port is not a number
		"8080:localhost:nope",             // forwarding port is not a number
		"/unix/is/no/go:/path/to/nowhere", // unix socket is unsupported
	}
	for _, tunnelStr := range invalids {
		tun, err := ParseTunnelArgument(tunnelStr, ssh.UnsetTunnel)
		if err == nil {
			t.Errorf("Parsed tunnel %v, want error", tun)
		}
	}
}
