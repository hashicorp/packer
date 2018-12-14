package compute

// VirtNICsClient defines a vritual nics client
type VirtNICsClient struct {
	ResourceClient
}

// VirtNICs returns a virtual nics client
func (c *Client) VirtNICs() *VirtNICsClient {
	return &VirtNICsClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: "Virtual NIC",
			ContainerPath:       "/network/v1/vnic/",
			ResourceRootPath:    "/network/v1/vnic",
		},
	}
}

// VirtualNIC defines the attributes in a virtual nic
type VirtualNIC struct {
	// Description of the object.
	Description string `json:"description"`
	// Fully Qualified Domain Name
	FQDN string `json:"name"`
	// MAC address of this VNIC.
	MACAddress string `json:"macAddress"`
	// The three-part name (/Compute-identity_domain/user/object) of the Virtual NIC.
	Name string
	// Tags associated with the object.
	Tags []string `json:"tags"`
	// True if the VNIC is of type "transit".
	TransitFlag bool `json:"transitFlag"`
	// Uniform Resource Identifier
	URI string `json:"uri"`
}

// GetVirtualNICInput Can only GET a virtual NIC, not update, create, or delete
type GetVirtualNICInput struct {
	// The three-part name (/Compute-identity_domain/user/object) of the Virtual NIC.
	// Required
	Name string `json:"name"`
}

// GetVirtualNIC returns the specified virtual nic
func (c *VirtNICsClient) GetVirtualNIC(input *GetVirtualNICInput) (*VirtualNIC, error) {
	var virtNIC VirtualNIC
	input.Name = c.getQualifiedName(input.Name)
	if err := c.getResource(input.Name, &virtNIC); err != nil {
		return nil, err
	}
	return c.success(&virtNIC)
}

func (c *VirtNICsClient) success(info *VirtualNIC) (*VirtualNIC, error) {
	info.Name = c.getUnqualifiedName(info.FQDN)
	return info, nil
}
