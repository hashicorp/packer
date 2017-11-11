package oci

import (
	"time"
)

// InstanceService enables communicating with the OCI compute API's instance
// related endpoints.
type InstanceService struct {
	client *baseClient
}

// NewInstanceService creates a new InstanceService for communicating with the
// OCI compute API's instance related endpoints.
func NewInstanceService(s *baseClient) *InstanceService {
	return &InstanceService{
		client: s.New().Path("instances/"),
	}
}

// Instance details a OCI compute instance.
type Instance struct {
	// The Availability Domain the instance is running in.
	AvailabilityDomain string `json:"availabilityDomain"`

	// The OCID of the compartment that contains the instance.
	CompartmentID string `json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	DisplayName string `json:"displayName,omitempty"`

	// The OCID of the instance.
	ID string `json:"id"`

	// The image used to boot the instance.
	ImageID string `json:"imageId,omitempty"`

	// The current state of the instance. Allowed values:
	//  - PROVISIONING
	//  - RUNNING
	//  - STARTING
	//  - STOPPING
	//  - STOPPED
	//  - CREATING_IMAGE
	//  - TERMINATING
	//  - TERMINATED
	LifecycleState string `json:"lifecycleState"`

	// Custom metadata that you provide.
	Metadata map[string]string `json:"metadata,omitempty"`

	// The region that contains the Availability Domain the instance is running in.
	Region string `json:"region"`

	// The shape of the instance. The shape determines the number of CPUs
	// and the amount of memory allocated to the instance.
	Shape string `json:"shape"`

	// The date and time the instance was created.
	TimeCreated time.Time `json:"timeCreated"`
}

// GetInstanceParams are the paramaters available when communicating with the
// GetInstance API endpoint.
type GetInstanceParams struct {
	ID string `url:"instanceId,omitempty"`
}

// Get returns a single Instance
func (s *InstanceService) Get(params *GetInstanceParams) (Instance, error) {
	instance := Instance{}
	e := &APIError{}

	_, err := s.client.New().Get(params.ID).Receive(&instance, e)
	err = firstError(err, e)

	return instance, err
}

// LaunchInstanceParams are the parameters available when communicating with
// the LunchInstance API endpoint.
type LaunchInstanceParams struct {
	AvailabilityDomain string            `json:"availabilityDomain,omitempty"`
	CompartmentID      string            `json:"compartmentId,omitempty"`
	DisplayName        string            `json:"displayName,omitempty"`
	ImageID            string            `json:"imageId,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	OPCiPXEScript      string            `json:"opcIpxeScript,omitempty"`
	Shape              string            `json:"shape,omitempty"`
	SubnetID           string            `json:"subnetId,omitempty"`
}

// Launch creates a new OCI compute instance. It does *not* wait for the
// instance to boot.
func (s *InstanceService) Launch(params *LaunchInstanceParams) (Instance, error) {
	instance := &Instance{}
	e := &APIError{}

	_, err := s.client.New().Post("").SetBody(params).Receive(instance, e)
	err = firstError(err, e)

	return *instance, err
}

// TerminateInstanceParams are the parameters available when communicating with
// the TerminateInstance API endpoint.
type TerminateInstanceParams struct {
	ID string `url:"instanceId,omitempty"`
}

// Terminate terminates a running OCI compute instance.
// instance to boot.
func (s *InstanceService) Terminate(params *TerminateInstanceParams) error {
	e := &APIError{}

	_, err := s.client.New().Delete(params.ID).SetBody(params).Receive(nil, e)
	err = firstError(err, e)

	return err
}

// GetResourceState GETs the LifecycleState of the given instance id.
func (s *InstanceService) GetResourceState(id string) (string, error) {
	instance, err := s.Get(&GetInstanceParams{ID: id})
	if err != nil {
		return "", err
	}
	return instance.LifecycleState, nil
}
