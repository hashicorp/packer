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
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
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
const APIVersion = "4"

// Server waits for a connection to this plugin and returns a Packer
// RPC server that you can use to register components and serve them.
func Server() (*packrpc.Server, error) {
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

	listener, err := serverListener(minPort, maxPort)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	// Output the address to stdout
	log.Printf("Plugin address: %s %s\n",
		listener.Addr().Network(), listener.Addr().String())
	fmt.Printf("%s|%s|%s\n",
		APIVersion,
		listener.Addr().Network(),
		listener.Addr().String())
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

func serverListener(minPort, maxPort int64) (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp(minPort, maxPort)
	}

	return serverListener_unix()
}

func serverListener_tcp(minPort, maxPort int64) (net.Listener, error) {
	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

func serverListener_unix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "packer-plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return net.Listen("unix", path)
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}
