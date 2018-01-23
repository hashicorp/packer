package oci

// ComputeClient is a client for the OCI Compute API.
type ComputeClient struct {
	BaseURL         string
	Instances       *InstanceService
	Images          *ImageService
	VNICAttachments *VNICAttachmentService
	VNICs           *VNICService
}

// NewComputeClient creates a new client for communicating with the OCI
// Compute API.
func NewComputeClient(s *baseClient) *ComputeClient {
	return &ComputeClient{
		Instances:       NewInstanceService(s),
		Images:          NewImageService(s),
		VNICAttachments: NewVNICAttachmentService(s),
		VNICs:           NewVNICService(s),
	}
}
