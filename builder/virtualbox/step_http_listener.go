package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"log"
	"math/rand"
	"net"
)

// This step finds an open port and binds the HTTP server to it.
//
// Uses:
//   config *config
//
// Produces:
//   http_listener net.Listener - The listener.
//   http_port int - The port the HTTP server started on.
type stepHTTPListener struct {
	l net.Listener
}

func (s *stepHTTPListener) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)

	var httpPort uint = 0
	if config.HTTPDir == "" {
		state["http_listener"] = nil
		state["http_port"] = httpPort
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

	state["http_listener"] = s.l
	state["http_port"] = httpPort
	return multistep.ActionContinue
}

func (s *stepHTTPListener) Cleanup(map[string]interface{}) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
