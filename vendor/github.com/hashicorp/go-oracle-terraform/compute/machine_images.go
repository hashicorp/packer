package compute

// MachineImagesClient is a client for the MachineImage functions of the Compute API.
type MachineImagesClient struct {
	ResourceClient
}

// MachineImages obtains an MachineImagesClient which can be used to access to the
// MachineImage functions of the Compute API
func (c *ComputeClient) MachineImages() *MachineImagesClient {
	return &MachineImagesClient{
		ResourceClient: ResourceClient{
			ComputeClient:       c,
			ResourceDescription: "MachineImage",
			ContainerPath:       "/machineimage/",
			ResourceRootPath:    "/machineimage",
		}}
}

// MahineImage describes an existing Machine Image.
type MachineImage struct {
	// account of the associated Object Storage Classic instance
	Account string `json:"account"`

	// Dictionary of attributes to be made available to the instance
	Attributes map[string]interface{} `json:"attributes"`

	// Last time when this image was audited
	Audited string `json:"audited"`

	// Describing the image
	Description string `json:"description"`

	// Description of the state of the machine image if there is an error
	ErrorReason string `json:"error_reason"`

	//  dictionary of hypervisor-specific attributes
	Hypervisor map[string]interface{} `json:"hypervisor"`

	// The format of the image
	ImageFormat string `json:"image_format"`

	// name of the machine image file uploaded to Object Storage Classic
	File string `json:"file"`

	// name of the machine image
	Name string `json:"name"`

	// Indicates that the image file is available in Object Storage Classic
	NoUpload bool `json:"no_upload"`

	// The OS platform of the image
	Platform string `json:"platform"`

	// Size values of the image file
	Sizes map[string]interface{} `json:"sizes"`

	// The state of the uploaded machine image
	State string `json:"state"`

	// Uniform Resource Identifier
	URI string `json:"uri"`
}

// CreateMachineImageInput defines an Image List to be created.
type CreateMachineImageInput struct {
	// account of the associated Object Storage Classic instance
	Account string `json:"account"`

	// Dictionary of attributes to be made available to the instance
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Describing the image
	Description string `json:"description,omitempty"`

	// name of the machine image file uploaded to Object Storage Classic
	File string `json:"file,omitempty"`

	// name of the machine image
	Name string `json:"name"`

	// Indicates that the image file is available in Object Storage Classic
	NoUpload bool `json:"no_upload"`

	// Size values of the image file
	Sizes map[string]interface{} `json:"sizes"`
}

// CreateMachineImage creates a new Machine Image with the given parameters.
func (c *MachineImagesClient) CreateMachineImage(createInput *CreateMachineImageInput) (*MachineImage, error) {
	var machineImage MachineImage

	// If `sizes` is not set then is mst be defaulted to {"total": 0}
	if createInput.Sizes == nil {
		createInput.Sizes = map[string]interface{}{"total": 0}
	}

	// `no_upload` must always be true
	createInput.NoUpload = true

	createInput.Name = c.getQualifiedName(createInput.Name)
	if err := c.createResource(createInput, &machineImage); err != nil {
		return nil, err
	}

	return c.success(&machineImage)
}

// DeleteMachineImageInput describes the MachineImage to delete
type DeleteMachineImageInput struct {
	// The name of the MachineImage
	Name string `json:name`
}

// DeleteMachineImage deletes the MachineImage with the given name.
func (c *MachineImagesClient) DeleteMachineImage(deleteInput *DeleteMachineImageInput) error {
	return c.deleteResource(deleteInput.Name)
}

// GetMachineList describes the MachineImage to get
type GetMachineImageInput struct {
	// account of the associated Object Storage Classic instance
	Account string `json:"account"`
	// The name of the Machine Image
	Name string `json:name`
}

// GetMachineImage retrieves the MachineImage with the given name.
func (c *MachineImagesClient) GetMachineImage(getInput *GetMachineImageInput) (*MachineImage, error) {
	getInput.Name = c.getQualifiedName(getInput.Name)

	var machineImage MachineImage
	if err := c.getResource(getInput.Name, &machineImage); err != nil {
		return nil, err
	}

	return c.success(&machineImage)
}

func (c *MachineImagesClient) success(result *MachineImage) (*MachineImage, error) {
	c.unqualify(&result.Name)
	return result, nil
}
