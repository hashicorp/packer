package compute

import "fmt"

const (
	imageListEntryDescription   = "image list entry"
	imageListEntryContainerPath = "/imagelist"
	imageListEntryResourcePath  = "/imagelist"
)

// ImageListEntriesClient specifies the parameters for an image list entries client
type ImageListEntriesClient struct {
	ResourceClient
}

// ImageListEntries returns an ImageListEntriesClient that can be used to access the
// necessary CRUD functions for Image List Entry's.
func (c *Client) ImageListEntries() *ImageListEntriesClient {
	return &ImageListEntriesClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: imageListEntryDescription,
			ContainerPath:       imageListEntryContainerPath,
			ResourceRootPath:    imageListEntryResourcePath,
		},
	}
}

// ImageListEntryInfo contains the exported fields necessary to hold all the information about an
// Image List Entry
type ImageListEntryInfo struct {
	// User-defined parameters, in JSON format, that can be passed to an instance of this machine
	// image when it is launched. This field can be used, for example, to specify the location of
	// a database server and login details. Instance metadata, including user-defined data is available
	// at http://192.0.0.192/ within an instance. See Retrieving User-Defined Instance Attributes in Using
	// Oracle Compute Cloud Service (IaaS).
	Attributes map[string]interface{} `json:"attributes"`
	// Name of the imagelist.
	Name string `json:"imagelist"`
	// A list of machine images.
	MachineImages []string `json:"machineimages"`
	// Uniform Resource Identifier for the Image List Entry
	URI string `json:"uri"`
	// Version number of these machineImages in the imagelist.
	Version int `json:"version"`
}

// CreateImageListEntryInput specifies the parameters needed to creat an image list entry
type CreateImageListEntryInput struct {
	// The name of the Image List
	Name string
	// User-defined parameters, in JSON format, that can be passed to an instance of this machine
	// image when it is launched. This field can be used, for example, to specify the location of
	// a database server and login details. Instance metadata, including user-defined data is
	//available at http://192.0.0.192/ within an instance. See Retrieving User-Defined Instance
	//Attributes in Using Oracle Compute Cloud Service (IaaS).
	// Optional
	Attributes map[string]interface{} `json:"attributes"`
	// A list of machine images.
	// Required
	MachineImages []string `json:"machineimages"`
	// The unique version of the entry in the image list.
	// Required
	Version int `json:"version"`
}

// CreateImageListEntry creates a new Image List Entry from an ImageListEntriesClient and an input struct.
// Returns a populated Info struct for the Image List Entry, and any errors
func (c *ImageListEntriesClient) CreateImageListEntry(input *CreateImageListEntryInput) (*ImageListEntryInfo, error) {
	c.updateClientPaths(input.Name, -1)
	var imageListEntryInfo ImageListEntryInfo
	if err := c.createResource(&input, &imageListEntryInfo); err != nil {
		return nil, err
	}
	return c.success(&imageListEntryInfo)
}

// GetImageListEntryInput details the parameters needed to retrive an image list entry
type GetImageListEntryInput struct {
	// The name of the Image List
	Name string
	// Version number of these machineImages in the imagelist.
	Version int
}

// GetImageListEntry returns a populated ImageListEntryInfo struct from an input struct
func (c *ImageListEntriesClient) GetImageListEntry(input *GetImageListEntryInput) (*ImageListEntryInfo, error) {
	c.updateClientPaths(input.Name, input.Version)
	var imageListEntryInfo ImageListEntryInfo
	if err := c.getResource("", &imageListEntryInfo); err != nil {
		return nil, err
	}
	return c.success(&imageListEntryInfo)
}

// DeleteImageListEntryInput details the parameters needed to delete an image list entry
type DeleteImageListEntryInput struct {
	// The name of the Image List
	Name string
	// Version number of these machineImages in the imagelist.
	Version int
}

// DeleteImageListEntry deletes the specified image list entry
func (c *ImageListEntriesClient) DeleteImageListEntry(input *DeleteImageListEntryInput) error {
	c.updateClientPaths(input.Name, input.Version)
	return c.deleteResource("")
}

func (c *ImageListEntriesClient) updateClientPaths(name string, version int) {
	var containerPath, resourcePath string
	name = c.getQualifiedName(name)
	containerPath = imageListEntryContainerPath + name + "/entry/"
	resourcePath = imageListEntryContainerPath + name + "/entry"
	if version != -1 {
		containerPath = fmt.Sprintf("%s%d", containerPath, version)
		resourcePath = fmt.Sprintf("%s/%d", resourcePath, version)
	}
	c.ContainerPath = containerPath
	c.ResourceRootPath = resourcePath
}

// Unqualifies any qualified fields in the IPNetworkInfo struct
func (c *ImageListEntriesClient) success(info *ImageListEntryInfo) (*ImageListEntryInfo, error) {
	c.unqualifyURL(&info.URI)
	return info, nil
}
