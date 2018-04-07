package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"

	"github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates and runs the HTTP server that is serving files from the
// directory specified by the 'http_directory` configuration parameter in the
// template.
//
// Uses:
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type StepHTTPServer struct {
	HTTPDir     string
	HTTPPortMin uint
	HTTPPortMax uint

	l net.Listener
}

func (s *StepHTTPServer) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var httpPort uint = 0
	if s.HTTPDir == "" {
		state.Put("http_port", httpPort)
		return multistep.ActionContinue
	}

	// Find an available TCP port for our HTTP server
	var httpAddr string
	portRange := int(s.HTTPPortMax - s.HTTPPortMin)
	for {
		var err error
		var offset uint = 0

		if portRange > 0 {
			// Intn will panic if portRange == 0, so we do a check.
			// Intn is from [0, n), so add 1 to make from [0, n]
			offset = uint(rand.Intn(portRange + 1))
		}

		httpPort = offset + s.HTTPPortMin
		httpAddr = fmt.Sprintf("0.0.0.0:%d", httpPort)
		log.Printf("Trying port: %d", httpPort)
		s.l, err = net.Listen("tcp", httpAddr)
		if err == nil {
			break
		}
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", httpPort))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(s.HTTPDir))
	server := &http.Server{Addr: httpAddr, Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", httpPort)
	SetHTTPPort(fmt.Sprintf("%d", httpPort))

	return multistep.ActionContinue
}

func SetHTTPPort(port string) error {
	return common.SetSharedState("port", port)
}

func SetHTTPIP(ip string) error {
	return common.SetSharedState("ip", ip)
}

func GetHTTPAddr() string {
	ip, err := common.RetrieveSharedState("ip")
	if err != nil {
		return ""
	}

	port, err := common.RetrieveSharedState("port")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", ip, port)
}

func (s *StepHTTPServer) Cleanup(multistep.StateBag) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
	common.RemoveSharedStateFile("port")
	common.RemoveSharedStateFile("ip")
}
