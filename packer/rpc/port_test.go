package rpc

import (
	"cgl.tideland.biz/asserts"
	"net"
	"strings"
	"testing"
)

func addrPort(address net.Addr) string {
	parts := strings.Split(address.String(), ":")
	return parts[len(parts) - 1]
}

func Test_netListenerInRange(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Open up port 10000 so that we take up a port
	L1000, err := net.Listen("tcp", ":10000")
	defer L1000.Close()
	assert.Nil(err, "should be able to bind to port 10000")

	// Verify it selects an open port
	L := netListenerInRange(10000, 10005)
	assert.NotNil(L, "should have a listener")
	assert.Equal(addrPort(L.Addr()), "10001", "should bind to open port")

	// Returns nil if there are no open ports
	L = netListenerInRange(10000, 10000)
	assert.Nil(L, "should not get a listener")
}
