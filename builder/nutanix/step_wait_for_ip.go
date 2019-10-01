package nutanix

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepWaitForIP struct {
	ClusterURL string
	Config     Config
	Timeout    time.Duration
}

// RetrieveReadyIP will query nutanix and wait for a COMPLETE status and identify an IP in the nic list
func RetrieveReadyIP(ctx context.Context, ipChan chan string, errChan chan error, req *http.Request) {
	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//Send request to create VM
	hc := http.Client{}
	var vmResponse *v3.VMIntentResponse

	for {
		resp, _ := hc.Do(req)
		if resp != nil {
			bodyText, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("An error occurred parsing the body: %s", err.Error())
				errChan <- err
				return
			}
			json.Unmarshal([]byte(bodyText), &vmResponse)
			log.Printf("Nutanix VM UUID: %s", *vmResponse.Metadata.UUID)
			if *vmResponse.Status.State == "COMPLETE" {
				j, _ := json.Marshal(vmResponse)
				log.Printf("Status is COMPLETE, retrieving IP from JSON: %s", j)
				if len(vmResponse.Spec.Resources.NicList) > 0 && len(vmResponse.Spec.Resources.NicList[0].IPEndpointList) > 0 {
					ipChan <- *(vmResponse.Spec.Resources.NicList[0].IPEndpointList)[0].IP
				} else {
					log.Printf("Ip endpoint not found yet. The NicList size is %d.", len(vmResponse.Spec.Resources.NicList))
				}
			} else if *vmResponse.Status.State == "ERROR" {
				var errTxt string
				for i := 0; i < len(vmResponse.Status.MessageList); i++ {
					errTxt = *(vmResponse.Status.MessageList)[i].Message
					log.Printf("Nutanix Error Message: %s", *(vmResponse.Status.MessageList)[i].Message)
					log.Printf("Nutanix Error Reason: %s", *(vmResponse.Status.MessageList)[i].Reason)
					log.Printf("Nutanix Error Details: %s", (vmResponse.Status.MessageList)[i].Details)
				}
				errChan <- errors.New(errTxt)
				return
			} else {
				log.Printf("Current status is: " + *vmResponse.Status.State)
			}
			defer resp.Body.Close()
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *stepWaitForIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmUUID := state.Get("vmUUID").(string)

	ui.Say("Retrieving VM status for uuid: " + vmUUID)

	ipChan := make(chan string)
	errChan := make(chan error)
	d := NewDriver(&s.Config.NutanixCluster, state)
	req, err := d.MakeRequest(http.MethodGet, "/vms/"+vmUUID, nil)
	if err != nil {
		ui.Error("Unable to create Nutanix VM request: " + err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	go RetrieveReadyIP(ctx, ipChan, errChan, req)

	shutdownTimer := time.After(s.Timeout)

	// loop http requests until status and IP are available or timeout
	for {
		select {
		case <-shutdownTimer:
			err := errors.New("Timeout occurred waiting for VM to be ready")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		case ip := <-ipChan:
			state.Put("ip", ip)
			ui.Say("IP for Nutanix device: " + ip)
			log.Println("Config Type is: ", s.Config.Config.Type)
			return multistep.ActionContinue
		case err := <-errChan:
			ui.Error("An error occurred retrieving the VM IP address: " + err.Error())
			log.Printf("Error returned: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *stepWaitForIP) Cleanup(state multistep.StateBag) {
}
