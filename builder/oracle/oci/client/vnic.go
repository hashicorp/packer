package oci

import (
	"time"
)

// VNICService enables communicating with the OCI compute API's VNICs
// endpoint.
type VNICService struct {
	client *baseClient
}

// NewVNICService creates a new VNICService for communicating with the
// OCI compute API's instance related endpoints.
func NewVNICService(s *baseClient) *VNICService {
	return &VNICService{client: s.New().Path("vnics/")}
}

// VNIC - a  virtual network interface card.
type VNIC struct {
	AvailabilityDomain string    `json:"availabilityDomain"`
	CompartmentID      string    `json:"compartmentId"`
	DisplayName        string    `json:"displayName,omitempty"`
	ID                 string    `json:"id"`
	LifecycleState     string    `json:"lifecycleState"`
	PrivateIP          string    `json:"privateIp"`
	PublicIP           string    `json:"publicIp"`
	SubnetID           string    `json:"subnetId"`
	TimeCreated        time.Time `json:"timeCreated"`
}

// GetVNICParams are the paramaters available when communicating with the
// ListVNICs API endpoint.
type GetVNICParams struct {
	ID string `url:"vnicId"`
}

// Get returns an individual VNIC.
func (s *VNICService) Get(params *GetVNICParams) (VNIC, error) {
	VNIC := &VNIC{}
	e := &APIError{}

	_, err := s.client.New().Get(params.ID).Receive(VNIC, e)
	err = firstError(err, e)

	return *VNIC, err
}
