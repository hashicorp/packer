package ssh

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/sdk-internals/communicator/ssh"
)

// ParseTunnelArgument parses an SSH tunneling argument compatible with the openssh client form.
// Valid formats:
// `port:host:hostport`
// NYI `[bind_address:]port:host:hostport`
func ParseTunnelArgument(forward string, direction ssh.TunnelDirection) (ssh.TunnelSpec, error) {
	parts := strings.SplitN(forward, ":", 2)
	if len(parts) != 2 {
		return ssh.TunnelSpec{}, fmt.Errorf("Error parsing tunnel '%s': %v", forward, parts)
	}
	listeningPort, forwardingAddr := parts[0], parts[1]

	_, sPort, err := net.SplitHostPort(forwardingAddr)
	if err != nil {
		return ssh.TunnelSpec{}, fmt.Errorf("Error parsing forwarding, must be a tcp address: %s", err)
	}
	_, err = strconv.Atoi(sPort)
	if err != nil {
		return ssh.TunnelSpec{}, fmt.Errorf("Error parsing forwarding port, must be a valid port: %s", err)
	}
	_, err = strconv.Atoi(listeningPort)
	if err != nil {
		return ssh.TunnelSpec{}, fmt.Errorf("Error parsing listening port, must be a valid port: %s", err)
	}

	return ssh.TunnelSpec{
		Direction:   direction,
		ForwardAddr: forwardingAddr,
		ForwardType: "tcp",
		ListenAddr:  fmt.Sprintf("localhost:%s", listeningPort),
		ListenType:  "tcp",
	}, nil
	// So we parsed all that, and are just going to ignore it now. We would
	// have used the information to set the type here.
}
