package nutanix

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepDestroyVM struct {
	Config *Config
}

func (s *stepDestroyVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if vmUUID, ok := state.GetOk("vmUUID"); ok {
		if vmUUID != "" {
			ui.Say("Cleaning up Nutanix VM.")
			time.Sleep(10 * time.Second)
			//Allow Insecure TLS
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			hc := http.Client{}
			d := NewDriver(&s.Config.NutanixCluster, state)
			req, err := d.MakeRequest(http.MethodDelete, "/vms/"+vmUUID.(string), nil)
			if err != nil {
				ui.Error("Unable to create Nutanix VM request: " + err.Error())
				return multistep.ActionHalt
			}
			resp, err := hc.Do(req)
			if err != nil {
				log.Printf("Call to Nutanix to create VM failed.")
				return multistep.ActionHalt
			} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
				ui.Say("Nutanix VM has been successfully deleted.")
			} else {
				ui.Error("An error occurred destroying the VM.")
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *stepDestroyVM) Cleanup(state multistep.StateBag) {
}
