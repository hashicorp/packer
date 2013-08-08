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
//   http_listener net.Listener
//   http_port int
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type stepHTTPServer struct{}

func (s *stepHTTPServer) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	httpListener := state["http_listener"].(net.Listener)
	httpPort := state["http_port"].(int)
	ui := state["ui"].(packer.Ui)

	if config.HTTPDir == "" {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", httpPort))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(config.HTTPDir))
	server := &http.Server{Handler: fileServer}
	go server.Serve(httpListener)

	// Save the address into the state so it can be accessed in the future
	state["http_port"] = httpPort

	return multistep.ActionContinue
}

func (s *stepHTTPServer) Cleanup(map[string]interface{}) {}
