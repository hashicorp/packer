// The plugin package provides the functionality to both expose a Packer
// plugin binary and to connect to an existing Packer plugin binary.
//
// Packer supports plugins in the form of self-contained external static
// Go binaries. These binaries behave in a certain way (enforced by this
// package) and are connected to in a certain way (also enforced by this
// package).
package plugin

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
)

const MagicCookieKey = "PACKER_PLUGIN_MAGIC_COOKIE"
const MagicCookieValue = "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2"

// This serves a single RPC connection on the given RPC server on
// a random port.
func serve(server *rpc.Server) (err error) {
	if os.Getenv(MagicCookieKey) != MagicCookieValue {
		return errors.New("Please do not execute plugins directly. Packer will execute these for you.")
	}

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	minPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MIN_PORT"), 10, 32)
	if err != nil {
		return
	}

	maxPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MAX_PORT"), 10, 32)
	if err != nil {
		return
	}

	log.Printf("Plugin minimum port: %d\n", minPort)
	log.Printf("Plugin maximum port: %d\n", maxPort)

	// Set the RPC port range
	packrpc.PortRange(int(minPort), int(maxPort))

	var address string
	var listener net.Listener
	for port := minPort; port <= maxPort; port++ {
		address = fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", address)
		if err != nil {
			if !strings.Contains(err.Error(), "address already in use") {
				// Not an address already in use error, return.
				return
			} else {
				// Address is in use, just try another
				err = nil
				continue
			}
		}

		break
	}

	defer listener.Close()

	// Output the address to stdout
	log.Printf("Plugin address: %s\n", address)
	fmt.Println(address)
	os.Stdout.Sync()

	// Accept a connection
	log.Println("Waiting for connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("Error accepting connection: %s\n", err.Error())
		return
	}

	// Serve a single connection
	log.Println("Serving a plugin connection...")
	server.ServeConn(conn)
	return
}

// Registers a signal handler to "swallow" interrupts so that the
// plugin isn't killed. The main host Packer process is responsible
// for killing the plugins when interrupted.
func swallowInterrupts() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch
		log.Println("Received interrupt signal. Ignoring.")
	}()
}

// Serves a builder from a plugin.
func ServeBuilder(builder packer.Builder) {
	log.Println("Preparing to serve a builder plugin...")

	server := rpc.NewServer()
	packrpc.RegisterBuilder(server, builder)

	swallowInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// Serves a command from a plugin.
func ServeCommand(command packer.Command) {
	log.Println("Preparing to serve a command plugin...")

	server := rpc.NewServer()
	packrpc.RegisterCommand(server, command)

	swallowInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// Serves a hook from a plugin.
func ServeHook(hook packer.Hook) {
	log.Println("Preparing to serve a hook plugin...")

	server := rpc.NewServer()
	packrpc.RegisterHook(server, hook)

	swallowInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// Serves a post-processor from a plugin.
func ServePostProcessor(p packer.PostProcessor) {
	log.Println("Preparing to serve a post-processor plugin...")

	server := rpc.NewServer()
	packrpc.RegisterPostProcessor(server, p)

	swallowInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// Serves a provisioner from a plugin.
func ServeProvisioner(p packer.Provisioner) {
	log.Println("Preparing to serve a provisioner plugin...")

	server := rpc.NewServer()
	packrpc.RegisterProvisioner(server, p)

	swallowInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
