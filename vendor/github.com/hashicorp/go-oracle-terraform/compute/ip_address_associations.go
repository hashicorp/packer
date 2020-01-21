package compute

const (
	iPAddressAssociationDescription   = "ip address association"
	iPAddressAssociationContainerPath = "/network/v1/ipassociation/"
	iPAddressAssociationResourcePath  = "/network/v1/ipassociation"
)

// IPAddressAssociationsClient details the parameters for an ip address association client
type IPAddressAssociationsClient struct {
	ResourceClient
}

// IPAddressAssociations returns an IPAddressAssociationsClient that can be used to access the
// necessary CRUD functions for IP Address Associations.
func (c *Client) IPAddressAssociations() *IPAddressAssociationsClient {
	return &IPAddressAssociationsClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: iPAddressAssociationDescription,
			ContainerPath:       iPAddressAssociationContainerPath,
			ResourceRootPath:    iPAddressAssociationResourcePath,
		},
	}
}

// IPAddressAssociationInfo contains the exported fields necessary to hold all the information about an
// IP Address Association
type IPAddressAssociationInfo struct {
	// Fully Qualified Domain Name
	FQDN string `json:"name"`
	// The name of the NAT IP address reservation.
	IPAddressReservation string `json:"ipAddressReservation"`
	// Name of the virtual NIC associated with this NAT IP reservation.
	Vnic string `json:"vnic"`
	// The name of the IP Address Association
	Name string
	// Description of the IP Address Association
	Description string `json:"description"`
	// Slice of tags associated with the IP Address Association
	Tags []string `json:"tags"`
	// Uniform Resource Identifier for the IP Address Association
	URI string `json:"uri"`
}

// CreateIPAddressAssociationInput details the attributes needed to create an ip address association
type CreateIPAddressAssociationInput struct {
	// The name of the IP Address Association to create. Object names can only contain alphanumeric,
	// underscore, dash, and period characters. Names are case-sensitive.
	// Required
	Name string `json:"name"`

	// The name of the NAT IP address reservation.
	// Optional
	IPAddressReservation string `json:"ipAddressReservation,omitempty"`

	// Name of the virtual NIC associated with this NAT IP reservation.
	// Optional
	Vnic string `json:"vnic,omitempty"`

	// Description of the IPAddressAssociation
	// Optional
	Description string `json:"description"`

	// String slice of tags to apply to the IP Address Association object
	// Optional
	Tags []string `json:"tags"`
}

// CreateIPAddressAssociation creates a new IP Address Association from an IPAddressAssociationsClient and an input struct.
// Returns a populated Info struct for the IP Address Association, and any errors
func (c *IPAddressAssociationsClient) CreateIPAddressAssociation(input *CreateIPAddressAssociationInput) (*IPAddressAssociationInfo, error) {
	input.Name = c.getQualifiedName(input.Name)
	input.IPAddressReservation = c.getQualifiedName(input.IPAddressReservation)
	input.Vnic = c.getQualifiedName(input.Vnic)

	var ipInfo IPAddressAssociationInfo
	if err := c.createResource(&input, &ipInfo); err != nil {
		return nil, err
	}

	return c.success(&ipInfo)
}

// GetIPAddressAssociationInput details the parameters needed to retrieve an ip address association
type GetIPAddressAssociationInput struct {
	// The name of the IP Address Association to query for. Case-sensitive
	// Required
	Name string `json:"name"`
}

// GetIPAddressAssociation returns a populated IPAddressAssociationInfo struct from an input struct
func (c *IPAddressAssociationsClient) GetIPAddressAssociation(input *GetIPAddressAssociationInput) (*IPAddressAssociationInfo, error) {
	input.Name = c.getQualifiedName(input.Name)

	var ipInfo IPAddressAssociationInfo
	if err := c.getResource(input.Name, &ipInfo); err != nil {
		return nil, err
	}

	return c.success(&ipInfo)
}

// DeleteIPAddressAssociationInput details the parameters neccessary to delete an ip address association
type DeleteIPAddressAssociationInput struct {
	// The name of the IP Address Association to query for. Case-sensitive
	// Required
	Name string `json:"name"`
}

// DeleteIPAddressAssociation deletes the specified ip address association
func (c *IPAddressAssociationsClient) DeleteIPAddressAssociation(input *DeleteIPAddressAssociationInput) error {
	return c.deleteResource(input.Name)
}

// Unqualifies any qualified fields in the IPAddressAssociationInfo struct
func (c *IPAddressAssociationsClient) success(info *IPAddressAssociationInfo) (*IPAddressAssociationInfo, error) {
	info.Name = c.getUnqualifiedName(info.FQDN)
	c.unqualify(&info.Vnic)
	c.unqualify(&info.IPAddressReservation)
	return info, nil
}
