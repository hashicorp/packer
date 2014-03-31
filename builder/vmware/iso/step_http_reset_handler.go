package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	vboxiso "github.com/mitchellh/packer/builder/virtualbox/iso"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/http"
	"strconv"
	"time"
)

type stepHTTPResetHandler struct {
	state multistep.StateBag
}

func (s *stepHTTPResetHandler) Run(state multistep.StateBag) multistep.StepAction {
	s.state = state
	http.Handle(vboxiso.RESET_URL, http.HandlerFunc(s.Reset))
	return multistep.ActionContinue
}

func (s *stepHTTPResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received URL %v", r.RequestURI)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	seconds, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)

	if err != nil {
		msg := fmt.Sprintf("Parameter 'seconds' is missing or not a number: %v", r.RequestURI)
		log.Printf(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if seconds < 0 || seconds > vboxiso.MAX_RESET_SECONDS {
		msg := fmt.Sprintf("Parameter 'seconds' must be between 0 and %d: %v", vboxiso.MAX_RESET_SECONDS, r.RequestURI)
		log.Printf(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	vmName := s.state.Get("vmx_path").(string)
	ui := s.state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Resetting %v in %d seconds", vmName, seconds))
	fmt.Fprintf(w, "Resetting %v in %d seconds", vmName, seconds)
	go s.reset(seconds)
}

func (s *stepHTTPResetHandler) reset(seconds int64) {
	time.Sleep(time.Duration(seconds) * time.Second)
	vmName := s.state.Get("vmx_path").(string)
	log.Printf("Resetting %s", vmName)
	driver := s.state.Get("driver").(vmwcommon.Driver)
	config := s.state.Get("config").(*config)
	err := driver.Reset(vmName, config.Headless)
	if err != nil {
		log.Printf("Error resetting %s: %s", vmName, err)
	} else {
		log.Printf("%s has been reset", vmName)
	}
}

func (stepHTTPResetHandler) Cleanup(multistep.StateBag) {}
