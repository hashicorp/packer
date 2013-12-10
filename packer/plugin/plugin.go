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
	"net/rpc"
	"os"
	"os/signal"
	"runtime"
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

// This serves a single RPC connection on the given RPC server on
// a random port.
func serve(server *rpc.Server) (err error) {
	log.Printf("Plugin build against Packer '%s'", packer.GitCommit)

	if os.Getenv(MagicCookieKey) != MagicCookieValue {
		return errors.New("Please do not execute plugins directly. Packer will execute these for you.")
	}

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	listener := packrpc.NetListener()
	defer listener.Close()

	fmt.Printf("%s|%s\n", APIVersion, listener.Addr())
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

// Registers a signal handler to swallow and count interrupts so that the
// plugin isn't killed. The main host Packer process is responsible
// for killing the plugins when interrupted.
func countInterrupts() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		for {
			<-ch
			newCount := atomic.AddInt32(&Interrupts, 1)
			log.Printf("Received interrupt signal (count: %d). Ignoring.", newCount)
		}
	}()
}

// Serves a builder from a plugin.
func ServeBuilder(builder packer.Builder) {
	log.Println("Preparing to serve a builder plugin...")

	server := rpc.NewServer()
	packrpc.RegisterBuilder(server, builder)

	countInterrupts()
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

	countInterrupts()
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

	countInterrupts()
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

	countInterrupts()
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

	countInterrupts()
	if err := serve(server); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// Tests whether or not the plugin was interrupted or not.
func Interrupted() bool {
	return atomic.LoadInt32(&Interrupts) > 0
}
