// The packer/plugin package provides the functionality required for writing
// Packer plugins in the form of static binaries that are then executed and
// run. It also contains the functions necessary to run these external plugins.
package plugin

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"os"
	packrpc "github.com/mitchellh/packer/packer/rpc"
)

// This serves the plugin by starting the RPC server and serving requests.
// This function never returns.
func serve(server *packrpc.Server) {
	// Start up the server
	server.Start()

	// Output the address to stdout
	fmt.Println(server.Address())
	os.Stdout.Sync()

	// Never return, wait on a channel that never gets a message
	<-make(chan bool)
}

func ServeCommand(command packer.Command) {
	server := packrpc.NewServer()
	server.RegisterCommand(command)
	serve(server)
}
