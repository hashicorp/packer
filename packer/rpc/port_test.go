package rpc

import (
	"net"
	"strings"
	"testing"
)

func addrPort(address net.Addr) string {
	parts := strings.Split(address.String(), ":")
	return parts[len(parts)-1]
}

func Test_netListenerInRange(t *testing.T) {
	// Open up port 10000 so that we take up a port
	L1000, err := net.Listen("tcp", "127.0.0.1:11000")
	defer L1000.Close()
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	if err == nil {
		// Verify it selects an open port
		L := netListenerInRange(11000, 11005)
		if L == nil {
			t.Fatal("L should not be nil")
		}
		if addrPort(L.Addr()) != "11001" {
			t.Fatalf("bad: %s", L.Addr())
		}

		// Returns nil if there are no open ports
		L = netListenerInRange(11000, 11000)
		if L != nil {
			t.Fatalf("bad: %#v", L)
		}
	}
}
