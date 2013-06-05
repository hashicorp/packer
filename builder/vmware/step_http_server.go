package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"net"
	"net/http"
)

// This step creates and runs the HTTP server that is serving the files
// specified by the 'http_files` configuration parameter in the template.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type stepHTTPServer struct{
	l net.Listener
}

func (s *stepHTTPServer) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	httpPort := 8080
	httpAddr := fmt.Sprintf(":%d", httpPort)

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", httpPort))

	// Start the TCP listener
	var err error
	s.l, err = net.Listen("tcp", httpAddr)
	if err != nil {
		ui.Error(fmt.Sprintf("Error starting HTTP server: %s", err))
		return multistep.ActionHalt
	}

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(config.HTTPDir))
	server := &http.Server{Addr: httpAddr, Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state["http_port"] = httpPort

	return multistep.ActionContinue
}

func (s *stepHTTPServer) Cleanup(map[string]interface{}) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
