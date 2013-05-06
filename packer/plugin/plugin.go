// The packer/plugin package provides the functionality required for writing
// Packer plugins in the form of static binaries that are then executed and
// run. It also contains the functions necessary to run these external plugins.
package plugin

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"net"
	"net/rpc"
	"os"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"strconv"
)

// This serves a single RPC connection on the given RPC server on
// a random port.
func serve(server *rpc.Server) (err error) {
	minPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MIN_PORT"), 10, 32)
	if err != nil {
		return
	}

	maxPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MAX_PORT"), 10, 32)
	if err != nil {
		return
	}

	var address string
	var listener net.Listener
	for port := minPort; port <= maxPort; port++ {
		address = fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", address)
		if err != nil {
			return
		}

		break
	}

	defer listener.Close()

	// Output the address to stdout
	fmt.Println(address)
	os.Stdout.Sync()

	// Accept a connection
	conn, err := listener.Accept()
	if err != nil {
		return
	}

	// Serve a single connection
	server.ServeConn(conn)
	return
}

// Serves a command from a plugin.
func ServeCommand(command packer.Command) {
	server := rpc.NewServer()
	packrpc.RegisterCommand(server, command)

	if err := serve(server); err != nil {
		panic(err)
	}
}
