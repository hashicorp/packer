package compute

import (
	"fmt"
	"path/filepath"
)

// IPAddressReservationsClient is a client to manage ip address reservation resources
type IPAddressReservationsClient struct {
	*ResourceClient
}

const (
	iPAddressReservationDescription   = "IP Address Reservation"
	iPAddressReservationContainerPath = "/network/v1/ipreservation/"
	iPAddressReservationResourcePath  = "/network/v1/ipreservation"
	iPAddressReservationQualifier     = "/oracle/public"
)

// IPAddressReservations returns an IPAddressReservationsClient to manage IP address reservation
// resources
func (c *Client) IPAddressReservations() *IPAddressReservationsClient {
	return &IPAddressReservationsClient{
		ResourceClient: &ResourceClient{
			Client:              c,
			ResourceDescription: iPAddressReservationDescription,
			ContainerPath:       iPAddressReservationContainerPath,
			ResourceRootPath:    iPAddressReservationResourcePath,
		},
	}
}

// IPAddressReservation describes an IP Address reservation
type IPAddressReservation struct {
	// Description of the IP Address Reservation
	Description string `json:"description"`

	// Fully Qualified Domain Name
	FQDN string `json:"name"`

	// Reserved NAT IPv4 address from the IP Address Pool
	IPAddress string `json:"ipAddress"`

	// Name of the IP Address pool to reserve the NAT IP from
	IPAddressPool string `json:"ipAddressPool"`

	// Name of the reservation
	Name string

	// Tags associated with the object
	Tags []string `json:"tags"`

	// Uniform Resource Identified for the reservation
	URI string `json:"uri"`
}

const (
	// PublicIPAddressPool - public-ippool
	PublicIPAddressPool = "public-ippool"
	// PrivateIPAddressPool - cloud-ippool
	PrivateIPAddressPool = "cloud-ippool"
)

// CreateIPAddressReservationInput defines input parameters to create an ip address reservation
type CreateIPAddressReservationInput struct {
	// Description of the IP Address Reservation
	// Optional
	Description string `json:"description"`

	// IP Address pool from which to reserve an IP Address.
	// Can be one of the following:
	//
	// 'public-ippool' - When you attach an IP Address from this pool to an instance, you enable
	//                   access between the public Internet and the instance
	// 'cloud-ippool' - When you attach an IP Address from this pool to an instance, the instance
	//                  can communicate privately with other Oracle Cloud Services
	// Optional
	IPAddressPool string `json:"ipAddressPool"`

	// The name of the reservation to create
	// Required
	Name string `json:"name"`

	// Tags to associate with the IP Reservation
	// Optional
	Tags []string `json:"tags"`
}

// CreateIPAddressReservation creates an IP Address reservation, and returns the info struct and any errors
func (c *IPAddressReservationsClient) CreateIPAddressReservation(input *CreateIPAddressReservationInput) (*IPAddressReservation, error) {
	var ipAddrRes IPAddressReservation
	// Qualify supplied name
	input.Name = c.getQualifiedName(input.Name)
	// Qualify supplied address pool if not nil
	if input.IPAddressPool != "" {
		input.IPAddressPool = c.qualifyIPAddressPool(input.IPAddressPool)
	}

	if err := c.createResource(input, &ipAddrRes); err != nil {
		return nil, err
	}

	return c.success(&ipAddrRes)
}

// GetIPAddressReservationInput details the parameters to retrieve information on an ip address reservation
type GetIPAddressReservationInput struct {
	// Name of the IP Reservation
	// Required
	Name string `json:"name"`
}

// GetIPAddressReservation returns an IP Address Reservation and any errors
func (c *IPAddressReservationsClient) GetIPAddressReservation(input *GetIPAddressReservationInput) (*IPAddressReservation, error) {
	var ipAddrRes IPAddressReservation

	input.Name = c.getQualifiedName(input.Name)
	if err := c.getResource(input.Name, &ipAddrRes); err != nil {
		return nil, err
	}

	return c.success(&ipAddrRes)
}

// UpdateIPAddressReservationInput details the parameters to update an IP Address reservation
type UpdateIPAddressReservationInput struct {
	// Description of the IP Address Reservation
	// Optional
	Description string `json:"description"`

	// IP Address pool from which to reserve an IP Address.
	// Can be one of the following:
	//
	// 'public-ippool' - When you attach an IP Address from this pool to an instance, you enable
	//                   access between the public Internet and the instance
	// 'cloud-ippool' - When you attach an IP Address from this pool to an instance, the instance
	//                  can communicate privately with other Oracle Cloud Services
	// Optional
	IPAddressPool string `json:"ipAddressPool"`

	// The name of the reservation to create
	// Required
	Name string `json:"name"`

	// Tags to associate with the IP Reservation
	// Optional
	Tags []string `json:"tags"`
}

// UpdateIPAddressReservation updates the specified ip address reservation
func (c *IPAddressReservationsClient) UpdateIPAddressReservation(input *UpdateIPAddressReservationInput) (*IPAddressReservation, error) {
	var ipAddrRes IPAddressReservation

	// Qualify supplied name
	input.Name = c.getQualifiedName(input.Name)
	// Qualify supplied address pool if not nil
	if input.IPAddressPool != "" {
		input.IPAddressPool = c.qualifyIPAddressPool(input.IPAddressPool)
	}

	if err := c.updateResource(input.Name, input, &ipAddrRes); err != nil {
		return nil, err
	}

	return c.success(&ipAddrRes)
}

// DeleteIPAddressReservationInput details the parameters to delete an IP Address Reservation
type DeleteIPAddressReservationInput struct {
	// The name of the reservation to delete
	Name string `json:"name"`
}

// DeleteIPAddressReservation deletes the specified ip address reservation
func (c *IPAddressReservationsClient) DeleteIPAddressReservation(input *DeleteIPAddressReservationInput) error {
	input.Name = c.getQualifiedName(input.Name)
	return c.deleteResource(input.Name)
}

func (c *IPAddressReservationsClient) success(result *IPAddressReservation) (*IPAddressReservation, error) {
	result.Name = c.getUnqualifiedName(result.FQDN)
	if result.IPAddressPool != "" {
		result.IPAddressPool = c.unqualifyIPAddressPool(result.IPAddressPool)
	}

	return result, nil
}

func (c *IPAddressReservationsClient) qualifyIPAddressPool(input string) string {
	// Add '/oracle/public/'
	return fmt.Sprintf("%s/%s", iPAddressReservationQualifier, input)
}

func (c *IPAddressReservationsClient) unqualifyIPAddressPool(input string) string {
	// Remove '/oracle/public/'
	return filepath.Base(input)
}
