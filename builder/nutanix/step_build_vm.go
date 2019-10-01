package nutanix

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// stepBuildVM is the default struct which contains the step's information
type stepBuildVM struct {
	ClusterURL string
	Config     *Config
}

// Run is the primary function to build the image
func (s *stepBuildVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	//Update UI
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Packer Builder VM on Nutanix Cluster.")

	// Setup HTTP request to Prism API

	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Setup JSON request to Prism
	vmRequest := &v3.VMIntentInput{
		Spec:     s.Config.Spec,
		Metadata: s.Config.Metadata,
	}

	hc := http.Client{}
	d := NewDriver(&s.Config.NutanixCluster, state)
	req, err := d.MakeRequest(http.MethodPost, "/vms", vmRequest)
	if err != nil {
		ui.Error("Unable to create Nutanix VM request: " + err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	resp, err := hc.Do(req)
	if err != nil {
		ui.Error("Call to Nutanix API to create VM failed: " + err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("Response body: %s", bodyText)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Failed request is of type ImageStatus
		var imageStatus *v3.ImageStatus
		json.Unmarshal([]byte(bodyText), &imageStatus)

		ui.Error("Nutanix VM request failed with response code: " + strconv.Itoa(resp.StatusCode))
		var errTxt string
		if imageStatus.State != nil && *imageStatus.State == "ERROR" {
			for i := 0; i < len(imageStatus.MessageList); i++ {
				errTxt = *(imageStatus.MessageList)[i].Message
				log.Printf("Nutanix Error Message: %s", *(imageStatus.MessageList)[i].Message)
				log.Printf("Nutanix Error Reason: %s", *(imageStatus.MessageList)[i].Reason)
				log.Printf("Nutanix Error Details: %s", (imageStatus.MessageList)[i].Details)
			}
		}
		state.Put("error", errors.New(errTxt))
		return multistep.ActionHalt
	}
	var vmResponse *v3.VMIntentResponse
	json.Unmarshal([]byte(bodyText), &vmResponse)

	log.Printf("Nutanix VM UUID: %s", *vmResponse.Metadata.UUID)
	state.Put("vmUUID", *vmResponse.Metadata.UUID)

	return multistep.ActionContinue
}

// Cleanup will tear down the VM once the build is complete
func (s *stepBuildVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	if vmUUID, ok := state.GetOk("vmUUID"); ok {
		if vmUUID != "" {
			ui.Say("Cleaning up Nutanix VM.")

			//Allow Insecure TLS
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			hc := http.Client{}
			d := NewDriver(&s.Config.NutanixCluster, state)
			req, err := d.MakeRequest(http.MethodDelete, "/vms/"+vmUUID.(string), nil)
			if err != nil {
				ui.Error("Unable to create Nutanix VM request: " + err.Error())
				return
			}
			resp, err := hc.Do(req)
			if err != nil {
				log.Printf("Call to Nutanix to create VM failed.")
				return
			} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
				ui.Say("Nutanix VM has been successfully deleted.")
			}
		}
	}
}
