package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
	"net/http"
)

// This step creates and runs the HTTP server that is serving files from the
// directory specified by the 'http_directory` configuration parameter in the
// template.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type stepHTTPServer struct {
	l net.Listener
}

func (s *stepHTTPServer) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	var httpPort uint = 0
	if config.HTTPDir == "" {
		state.Put("http_port", httpPort)
		return multistep.ActionContinue
	}

	// Find an available TCP port for our HTTP server
	var httpAddr string
	portRange := int(config.HTTPPortMax - config.HTTPPortMin)
	for {
		var err error
		var offset uint = 0

		if portRange > 0 {
			// Intn will panic if portRange == 0, so we do a check.
			offset = uint(rand.Intn(portRange))
		}

		httpPort = offset + config.HTTPPortMin
		httpAddr = fmt.Sprintf(":%d", httpPort)
		log.Printf("Trying port: %d", httpPort)
		s.l, err = net.Listen("tcp", httpAddr)
		if err == nil {
			break
		}
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", httpPort))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(config.HTTPDir))
	server := &http.Server{Addr: httpAddr, Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", httpPort)

	return multistep.ActionContinue
}

func (s *stepHTTPServer) Cleanup(multistep.StateBag) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
