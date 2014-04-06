package iso

import (
	"github.com/mitchellh/multistep"
	"log"
	"net/http"
)

type stepHTTPFileHandler struct{}

func (s *stepHTTPFileHandler) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	fileServer := http.FileServer(http.Dir(config.HTTPDir))
	fileHander := func(w http.ResponseWriter, r *http.Request) {
		lw := &loggedResponse{ResponseWriter: w}
		fileServer.ServeHTTP(lw, r)
		log.Printf("Received HTTP request: [%s] %s %s %d", r.RemoteAddr, r.Method, r.RequestURI, lw.statusCode)
	}
	http.Handle("/", http.HandlerFunc(fileHander))
	return multistep.ActionContinue
}

type loggedResponse struct {
	http.ResponseWriter
	statusCode int
}

func (l *loggedResponse) WriteHeader(statusCode int) {
	l.statusCode = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}

func (stepHTTPFileHandler) Cleanup(multistep.StateBag) {}
