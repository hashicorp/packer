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
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
)

// This is a count of the number of interrupts the process has received.
// This is updated with sync/atomic whenever a SIGINT is received and can
// be checked by the plugin safely to take action.
var Interrupts int32 = 0

const MagicCookieKey = "PACKER_PLUGIN_MAGIC_COOKIE"
const MagicCookieValue = "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2"

// The APIVersion is outputted along with the RPC address. The plugin
// client validates this API version and will show an error if it doesn't
// know how to speak it.
const APIVersion = "1"

// Server waits for a connection to this plugin and returns a Packer
// RPC server that you can use to register components and serve them.
func Server() (*packrpc.Server, error) {
	log.Printf("Plugin build against Packer '%s'", packer.GitCommit)

	if os.Getenv(MagicCookieKey) != MagicCookieValue {
		return nil, errors.New(
			"Please do not execute plugins directly. Packer will execute these for you.")
	}

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	minPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MIN_PORT"), 10, 32)
	if err != nil {
		return nil, err
	}

	maxPort, err := strconv.ParseInt(os.Getenv("PACKER_PLUGIN_MAX_PORT"), 10, 32)
	if err != nil {
		return nil, err
	}

	log.Printf("Plugin minimum port: %d\n", minPort)
	log.Printf("Plugin maximum port: %d\n", maxPort)

	// Set the RPC port range
	packrpc.PortRange(int(minPort), int(maxPort))

	var address string
	var listener net.Listener
	for port := minPort; port <= maxPort; port++ {
		address = fmt.Sprintf("127.0.0.1:%d", port)
		listener, err = net.Listen("tcp", address)
		if err != nil {
			err = nil
			continue
		}

		break
	}
	defer listener.Close()

	// Output the address to stdout
	log.Printf("Plugin address: %s\n", address)
	fmt.Printf("%s|%s\n", APIVersion, address)
	os.Stdout.Sync()

	// Accept a connection
	log.Println("Waiting for connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("Error accepting connection: %s\n", err.Error())
		return nil, err
	}

	// Eat the interrupts
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		var count int32 = 0
		for {
			<-ch
			newCount := atomic.AddInt32(&count, 1)
			log.Printf("Received interrupt signal (count: %d). Ignoring.", newCount)
		}
	}()

	// Serve a single connection
	log.Println("Serving a plugin connection...")
	return packrpc.NewServer(conn), nil
}
