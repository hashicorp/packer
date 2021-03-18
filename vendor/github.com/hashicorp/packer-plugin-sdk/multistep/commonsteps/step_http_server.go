package commonsteps

import (
	"context"
	"fmt"
	"log"

	"net/http"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step creates and runs the HTTP server that is serving files from the
// directory specified by the 'http_directory` configuration parameter in the
// template.
//
// Uses:
//   ui     packersdk.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type StepHTTPServer struct {
	HTTPDir     string
	HTTPContent map[string]string
	HTTPPortMin int
	HTTPPortMax int
	HTTPAddress string

	l *net.Listener
}

type MapServer map[string]string

func (s MapServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	content, found := s[r.URL.Path]
	if !found {
		// TODO: this will be displayed on stdout, here we could implement a
		// "did you mean" for helps.
		http.Error(w, fmt.Sprintf("File %s not found", r.URL.Path), http.StatusNotFound)
		return
	}

	if _, err := w.Write([]byte(content)); err != nil {
		// log err in case the file couldn't be 100% transferred for example.
		log.Printf("http_content serve error: %w", err)
	}
}

func (s *StepHTTPServer) Handler() http.Handler {
	if s.HTTPDir != "" {
		return http.FileServer(http.Dir(s.HTTPDir))
	}

	return MapServer(s.HTTPContent)
}

func (s *StepHTTPServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.HTTPDir == "" && len(s.HTTPContent) == 0 {
		state.Put("http_port", 0)
		return multistep.ActionContinue
	}

	// Find an available TCP port for our HTTP server
	var httpAddr string
	var err error
	s.l, err = net.ListenRangeConfig{
		Min:     s.HTTPPortMin,
		Max:     s.HTTPPortMax,
		Addr:    s.HTTPAddress,
		Network: "tcp",
	}.Listen(ctx)

	if err != nil {
		err := fmt.Errorf("Error finding port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", s.l.Port))

	// Start the HTTP server and run it in the background
	server := &http.Server{Addr: httpAddr, Handler: s.Handler()}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", s.l.Port)

	return multistep.ActionContinue
}

func (s *StepHTTPServer) Cleanup(multistep.StateBag) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
