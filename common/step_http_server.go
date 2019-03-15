package common

import (
	"context"
	"fmt"

	"net/http"

	"github.com/hashicorp/packer/common/net"
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
	HTTPPortMin int
	HTTPPortMax int

	l *net.Listener
}

func (s *StepHTTPServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.HTTPDir == "" {
		state.Put("http_port", uint(0))
		return multistep.ActionContinue
	}

	// Find an available TCP port for our HTTP server
	var httpAddr string
	var err error
	s.l, err = net.ListenRangeConfig{
		Min:     s.HTTPPortMin,
		Max:     s.HTTPPortMax,
		Addr:    "0.0.0.0",
		Network: "tcp",
	}.Listen(ctx)

	if err != nil {
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", s.l.Port))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(s.HTTPDir))
	server := &http.Server{Addr: httpAddr, Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", s.l.Port)
	SetHTTPPort(fmt.Sprintf("%d", s.l.Port))

	return multistep.ActionContinue
}

func SetHTTPPort(port string) error {
	return common.SetSharedState("port", port, "")
}

func SetHTTPIP(ip string) error {
	return common.SetSharedState("ip", ip, "")
}

func GetHTTPAddr() string {
	ip, err := common.RetrieveSharedState("ip", "")
	if err != nil {
		return ""
	}

	port, err := common.RetrieveSharedState("port", "")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", ip, port)
}

func GetHTTPPort() string {
	port, err := common.RetrieveSharedState("port", "")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", port)
}

func GetHTTPIP() string {
	ip, err := common.RetrieveSharedState("ip", "")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", ip)
}

func (s *StepHTTPServer) Cleanup(multistep.StateBag) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
	common.RemoveSharedStateFile("port", "")
	common.RemoveSharedStateFile("ip", "")
}
