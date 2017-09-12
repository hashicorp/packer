package oci

import (
	"time"
)

// ImageService enables communicating with the OCI compute API's instance
// related endpoints.
type ImageService struct {
	client *baseClient
}

// NewImageService creates a new ImageService for communicating with the
// OCI compute API's instance related endpoints.
func NewImageService(s *baseClient) *ImageService {
	return &ImageService{
		client: s.New().Path("images/"),
	}
}

// Image details a OCI boot disk image.
type Image struct {
	// The OCID of the image originally used to launch the instance.
	BaseImageID string `json:"baseImageId,omitempty"`

	// The OCID of the compartment containing the instance you want to use
	// as the basis for the image.
	CompartmentID string `json:"compartmentId"`

	// Whether instances launched with this image can be used to create new
	// images.
	CreateImageAllowed bool `json:"createImageAllowed"`

	// A user-friendly name for the image. It does not have to be unique,
	// and it's changeable. You cannot use an Oracle-provided image name
	// as a custom image name.
	DisplayName string `json:"displayName,omitempty"`

	// The OCID of the image.
	ID string `json:"id"`

	// Current state of the image. Allowed values are:
	//  - PROVISIONING
	//  - AVAILABLE
	//  - DISABLED
	//  - DELETED
	LifecycleState string `json:"lifecycleState"`

	// The image's operating system (e.g. Oracle Linux).
	OperatingSystem string `json:"operatingSystem"`

	// The image's operating system version (e.g. 7.2).
	OperatingSystemVersion string `json:"operatingSystemVersion"`

	// The date and time the image was created.
	TimeCreated time.Time `json:"timeCreated"`
}

// GetImageParams are the paramaters available when communicating with the
// GetImage API endpoint.
type GetImageParams struct {
	ID string `url:"imageId"`
}

// Get returns a single Image
func (s *ImageService) Get(params *GetImageParams) (Image, error) {
	image := Image{}
	e := &APIError{}

	_, err := s.client.New().Get(params.ID).Receive(&image, e)
	err = firstError(err, e)

	return image, err
}

// CreateImageParams are the parameters available when communicating with
// the CreateImage API endpoint.
type CreateImageParams struct {
	CompartmentID string `json:"compartmentId"`
	DisplayName   string `json:"displayName,omitempty"`
	InstanceID    string `json:"instanceId"`
}

// Create creates a new custom image based on a running compute instance. It
// does *not* wait for the imaging process to finish.
func (s *ImageService) Create(params *CreateImageParams) (Image, error) {
	image := Image{}
	e := &APIError{}

	_, err := s.client.New().Post("").SetBody(params).Receive(&image, &e)
	err = firstError(err, e)

	return image, err
}

// GetResourceState GETs the LifecycleState of the given image id.
func (s *ImageService) GetResourceState(id string) (string, error) {
	image, err := s.Get(&GetImageParams{ID: id})
	if err != nil {
		return "", err
	}
	return image.LifecycleState, nil

}

// DeleteImageParams are the parameters available when communicating with
// the DeleteImage API endpoint.
type DeleteImageParams struct {
	ID string `url:"imageId"`
}

// Delete deletes an existing custom image.
// NOTE: Deleting an image results in the API endpoint returning 404 on
// subsequent calls. As such deletion can't be waited on with a Waiter.
func (s *ImageService) Delete(params *DeleteImageParams) error {
	e := &APIError{}

	_, err := s.client.New().Delete(params.ID).SetBody(params).Receive(nil, e)
	err = firstError(err, e)

	return err
}
