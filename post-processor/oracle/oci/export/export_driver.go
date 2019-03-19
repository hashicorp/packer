package export

import (
	"context"
	"fmt"
	"time"

	core "github.com/oracle/oci-go-sdk/core"
	os "github.com/oracle/oci-go-sdk/objectstorage"
)

// driverOCI implements the Driver interface and communicates with Oracle
// OCI.

type ExportClient struct {
	computeClient core.ComputeClient
	osClient      os.ObjectStorageClient
	cfg           *Config
}

// NewDriverOCI Creates a new driverOCI with a connected compute client and a connected vcn client.
func NewExportClient(cfg *Config) (*ExportClient, error) {
	coreClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return nil, err
	}
	osclient, err := os.NewObjectStorageClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return nil, err
	}

	return &ExportClient{
		computeClient: coreClient,
		osClient:      osclient,
		cfg:           cfg,
	}, nil
}

// CreateInstance creates a new compute instance.
func (d *ExportClient) ExportImage(imageid string) (string, error) {
	Namespace, err := d.getNamespace()
	if err != nil {
		return "", err
	}
	imageDetails := core.ExportImageViaObjectStorageTupleDetails{
		BucketName:    &d.cfg.BucketName,
		ObjectName:    &d.cfg.ImageName,
		NamespaceName: &Namespace,
	}

	exportImageRequest := core.ExportImageRequest{
		ImageId:            &imageid,
		ExportImageDetails: imageDetails,
	}

	response, err := d.computeClient.ExportImage(context.TODO(), exportImageRequest)

	if err != nil {
		return "", err
	}
	err = d.WaitForImageExport(context.TODO(), imageid)
	if err != nil {
		return "", err
	}

	return response.Image.String(), nil
}

func (d *ExportClient) getNamespace() (string, error) {
	request := os.GetNamespaceRequest{}
	r, err := d.osClient.GetNamespace(context.TODO(), request)
	if err != nil {
		return "", err
	}
	fmt.Println("get namespace")
	return *r.Value, nil
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *ExportClient) WaitForImageExport(ctx context.Context, id string) error {
	return waitForResourceToReachState(
		func(string) (string, error) {
			image, err := d.computeClient.GetImage(ctx, core.GetImageRequest{ImageId: &id})
			if err != nil {
				return "", err
			}
			return string(image.LifecycleState), nil
		},
		id,
		[]string{"EXPORTING"},
		"AVAILABLE",
		0,             //Unlimited Retries
		5*time.Second, //5 second wait between retries
	)
}

// WaitForResourceToReachState checks the response of a request through a
// polled get and waits until the desired state or until the max retried has
// been reached.
func waitForResourceToReachState(getResourceState func(string) (string, error), id string, waitStates []string, terminalState string, maxRetries int, waitDuration time.Duration) error {
	for i := 0; maxRetries == 0 || i < maxRetries; i++ {
		state, err := getResourceState(id)
		if err != nil {
			return err
		}

		if stringSliceContains(waitStates, state) {
			time.Sleep(waitDuration)
			continue
		} else if state == terminalState {
			return nil
		}
		return fmt.Errorf("Unexpected resource state %q, expecting a waiting state %s or terminal state  %q ", state, waitStates, terminalState)
	}
	return fmt.Errorf("Maximum number of retries (%d) exceeded; resource did not reach state %q", maxRetries, terminalState)
}

// stringSliceContains loops through a slice of strings returning a boolean
// based on whether a given value is contained in the slice.
func stringSliceContains(slice []string, value string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}
	return false
}
