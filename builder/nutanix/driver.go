package nutanix

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	nutanixcommon "github.com/hashicorp/packer/builder/nutanix/common"
	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// PrismAPIVersion sets the default Nutanix API version
const PrismAPIVersion int = 3

// PrismAPIURI is a constant URI path for nutanix REST service
const PrismAPIURI string = "/api/nutanix/v"

// Driver contains the params for the API calls to cloud
type Driver struct {
	cluster *nutanixcommon.NutanixCluster
	state   multistep.StateBag
}

// NewDriver creates a new driver
func NewDriver(cluster *nutanixcommon.NutanixCluster, state multistep.StateBag) *Driver {
	driver := &Driver{
		cluster: cluster,
		state:   state,
	}
	return driver
}

// MakeRequest generates a basic request to the API
func (d *Driver) MakeRequest(httpMethod, uri string, s interface{}) (*http.Request, error) {
	jsonReq, err := json.Marshal(s)
	if err != nil {
		log.Printf("Error creating JSON: %s", err)
		return nil, err
	}
	if jsonReq != nil {
		log.Printf("Nutanix Request JSON: %s", jsonReq)
	}

	//Initialize local vars
	nutanixAPIURL := fmt.Sprintf("%s%s%d", d.cluster.ClusterURL, PrismAPIURI, PrismAPIVersion)
	log.Printf("Nutanix URL: %s", nutanixAPIURL+uri)

	// Setup HTTP Request to Create Temporary VM
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest(httpMethod, nutanixAPIURL+uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		//ui.Error("An issue occurred creating temporary Nutanix VM")
		log.Printf("Error creating JSON: %s", err)
		return nil, err
	}
	log.Printf("Cluster User: %s", *d.cluster.ClusterUsername)
	req.SetBasicAuth(*d.cluster.ClusterUsername, *d.cluster.ClusterPassword)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (d *Driver) retrieveVM(ctx context.Context, vmUUID string, vmChan chan *v3.VMIntentResponse, errChan chan error) {
	req, err := d.MakeRequest(http.MethodGet, "/vms/"+vmUUID, nil)
	if err != nil {
		errChan <- errors.New(err.Error())
		return
	}

	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//Send request to create VM
	hc := http.Client{}
	var vmResponse *v3.VMIntentResponse

	for {
		resp, err := hc.Do(req)
		if err != nil {
			errTxt := "Call to Nutanix to create VM failed."
			log.Printf("Call to Nutanix to create VM failed: %s", err)
			errChan <- errors.New(errTxt)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errTxt := "Reading VM response failed."
			log.Printf("Reading VM response failed: %s", err)
			errChan <- errors.New(errTxt)
			return
		}
		defer resp.Body.Close()
		json.Unmarshal([]byte(bodyText), &vmResponse)
		log.Printf("Nutanix VM UUID: %s", *vmResponse.Metadata.UUID)
		if *vmResponse.Status.State == "COMPLETE" {
			vmChan <- vmResponse
			return
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
		time.Sleep(5 * time.Second)
	}
}

// retrieveTaskResult calls the API service and based on the UUID, retrieves the result
func (d *Driver) retrieveTaskResult(ctx context.Context, taskUUID string, taskChan chan *v3.TasksResponse, errChan chan error) {
	req, err := d.MakeRequest(http.MethodGet, "/tasks/"+taskUUID, nil)
	if err != nil {
		errChan <- errors.New(err.Error())
		return
	}

	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//Send request to retrieve the task
	hc := http.Client{}
	var taskResponse *v3.TasksResponse

	for {
		resp, err := hc.Do(req)
		if err != nil {
			log.Printf("Call to Nutanix to retrieve Tasks failed: %s", err)
			errChan <- errors.New("Call to Nutanix to retrieve Tasks failed")
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errTxt := "Reading task response failed."
			log.Printf("Reading task response failed: %s", err)
			errChan <- errors.New(errTxt)
			return
		}
		defer resp.Body.Close()
		json.Unmarshal([]byte(bodyText), &taskResponse)
		log.Printf("Nutanix Task UUID: %s", taskUUID)
		if *taskResponse.Status == "SUCCEEDED" {
			taskChan <- taskResponse
			return
		} else if *taskResponse.Status == "ERRORED" {
			errChan <- errors.New("The task failed to complete")
			return
		} else {
			log.Printf("Current status is: " + *taskResponse.Status)
		}
		time.Sleep(5 * time.Second)
	}
}

// RetrieveReadyVM calls the API and retrieves the details once it is ready
func (d *Driver) RetrieveReadyVM(ctx context.Context, timeout time.Duration) (*v3.VMIntentResponse, error) {
	ui := d.state.Get("ui").(packer.Ui)
	vmUUID := d.state.Get("vmUUID").(string)

	vmChan := make(chan *v3.VMIntentResponse)
	errChan := make(chan error)

	go d.retrieveVM(ctx, vmUUID, vmChan, errChan)

	shutdownTimer := time.After(timeout)

	// loop http requests until status and IP are available or timeout
	for {
		select {
		case <-shutdownTimer:
			errTxt := "Timeout occurred waiting for VM to be ready."
			ui.Error(errTxt)
			return nil, errors.New(errTxt)
		case vm := <-vmChan:
			return vm, nil
		case err := <-errChan:
			ui.Error("An error occurred retrieving the VM: " + err.Error())
			log.Printf("Error returned: %s", err)
			return nil, err
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// UpdateVM will update the virtual machine
func (d *Driver) UpdateVM(ctx context.Context, vmRequest *v3.VMIntentInput) (*v3.VMIntentResponse, error) {
	ui := d.state.Get("ui").(packer.Ui)
	vmUUID := d.state.Get("vmUUID").(string)

	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	hc := http.Client{}
	req, err := d.MakeRequest(http.MethodPut, "/vms/"+vmUUID, vmRequest)
	if err != nil {
		ui.Error("Unable to create Nutanix VM request: " + err.Error())
		return nil, err
	}
	resp, err := hc.Do(req)
	if err != nil {
		ui.Error("Call to Nutanix to create VM failed with response code: " + strconv.Itoa(resp.StatusCode))
		log.Printf("Call to Nutanix to create VM failed: %s", err)
		return nil, err
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("Response body: %s", bodyText)
	var vmResponse *v3.VMIntentResponse
	json.Unmarshal([]byte(bodyText), &vmResponse)

	if resp.StatusCode != http.StatusAccepted {
		ui.Error("Nutanix VM request failed with response code: " + strconv.Itoa(resp.StatusCode))
		if *vmResponse.Status.State == "ERROR" {
			var errTxt string
			for i := 0; i < len(vmResponse.Status.MessageList); i++ {
				errTxt = *(vmResponse.Status.MessageList)[i].Message
				log.Printf("Nutanix Error Message: %s", *(vmResponse.Status.MessageList)[i].Message)
				log.Printf("Nutanix Error Reason: %s", *(vmResponse.Status.MessageList)[i].Reason)
				log.Printf("Nutanix Error Details: %s", (vmResponse.Status.MessageList)[i].Details)
			}
			return nil, errors.New(errTxt)
		}
	}
	return vmResponse, nil
}

// SaveVMDisk will call the API and save the image based upon the vm disk
func (d *Driver) SaveVMDisk(ctx context.Context, imageIntentInput *v3.ImageIntentInput) (*v3.VMIntentResponse, error) {
	ui := d.state.Get("ui").(packer.Ui)

	//Allow Insecure TLS
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	hc := http.Client{}
	req, err := d.MakeRequest(http.MethodPost, "/images", imageIntentInput)
	if err != nil {
		ui.Error("Unable to create Nutanix VM request: " + err.Error())
		return nil, err
	}
	resp, err := hc.Do(req)
	if err != nil {
		ui.Error("Call to Nutanix to save VM DISK failed with response code: " + strconv.Itoa(resp.StatusCode))
		log.Printf("Call to Nutanix to save VM DISK failed: %s", err)
		return nil, err
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("Response body: %s", bodyText)

	if resp.StatusCode != http.StatusAccepted {
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
		return nil, errors.New(errTxt)
	}
	var vmResponse *v3.VMIntentResponse
	json.Unmarshal([]byte(bodyText), &vmResponse)

	return vmResponse, nil
}

// RetrieveTask calls the API and returns the task details
func (d *Driver) RetrieveTask(ctx context.Context, taskUUID string) (*v3.TasksResponse, error) {
	ui := d.state.Get("ui").(packer.Ui)

	taskChan := make(chan *v3.TasksResponse)
	errChan := make(chan error)

	go d.retrieveTaskResult(ctx, taskUUID, taskChan, errChan)

	shutdownTimer := time.After(1 * time.Minute)

	// loop http requests until status and IP are available or timeout
	for {
		select {
		case <-shutdownTimer:
			errTxt := "Timeout occurred waiting for task to complete."
			ui.Error(errTxt)
			return nil, errors.New(errTxt)
		case task := <-taskChan:
			ui.Message("Nutanix task completed successfully.")
			return task, nil
		case err := <-errChan:
			ui.Error("An error occurred retrieving the task: " + err.Error())
			log.Printf("Error returned: %s", err)
			return nil, err
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
