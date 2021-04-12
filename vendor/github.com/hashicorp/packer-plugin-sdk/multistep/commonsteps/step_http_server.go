package commonsteps

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sort"

	"github.com/hashicorp/packer-plugin-sdk/didyoumean"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func HTTPServerFromHTTPConfig(cfg *HTTPConfig) *StepHTTPServer {
	return &StepHTTPServer{
		HTTPDir:     cfg.HTTPDir,
		HTTPContent: cfg.HTTPContent,
		HTTPPortMin: cfg.HTTPPortMin,
		HTTPPortMax: cfg.HTTPPortMax,
		HTTPAddress: cfg.HTTPAddress,
	}
}

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

func (s *StepHTTPServer) Handler() http.Handler {
	if s.HTTPDir != "" {
		return http.FileServer(http.Dir(s.HTTPDir))
	}

	return MapServer(s.HTTPContent)
}

type MapServer map[string]string

func (s MapServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := path.Clean(r.URL.Path)
	content, found := s[path]
	if !found {
		paths := make([]string, 0, len(s))
		for k := range s {
			paths = append(paths, k)
		}
		sort.Strings(paths)
		err := fmt.Sprintf("%s not found.", path)
		if sug := didyoumean.NameSuggestion(path, paths); sug != "" {
			err += fmt.Sprintf(" Did you mean %q?", sug)
		}

		http.Error(w, err, http.StatusNotFound)
		return
	}

	if _, err := w.Write([]byte(content)); err != nil {
		// log err in case the file couldn't be 100% transferred for example.
		log.Printf("http_content serve error: %v", err)
	}
}

func (s *StepHTTPServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.HTTPDir == "" && len(s.HTTPContent) == 0 {
		state.Put("http_port", 0)
		return multistep.ActionContinue
	}

	if s.HTTPDir != "" {
		if _, err := os.Stat(s.HTTPDir); err != nil {
			err := fmt.Errorf("Error finding %q: %s", s.HTTPDir, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Find an available TCP port for our HTTP server
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
	server := &http.Server{Addr: "", Handler: s.Handler()}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", s.l.Port)

	return multistep.ActionContinue
}

func (s *StepHTTPServer) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		ui := state.Get("ui").(packersdk.Ui)

		// Close the listener so that the HTTP server stops
		if err := s.l.Close(); err != nil {
			err = fmt.Errorf("Failed closing http server on port %d: %w", s.l.Port, err)
			ui.Error(err.Error())
			// Here this error should be shown to the UI but it won't
			// specifically stop Packer from terminating successfully. It could
			// cause a "Listen leak" if it happenned a lot. Though Listen will
			// try other ports if one is already used. In the case we want to
			// Listen on only one port, the next Listen call could fail or be
			// longer than expected.
		}
	}
}
