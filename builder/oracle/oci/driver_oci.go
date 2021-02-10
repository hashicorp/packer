package oci

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	core "github.com/oracle/oci-go-sdk/core"
)

// driverOCI implements the Driver interface and communicates with Oracle
// OCI.
type driverOCI struct {
	computeClient core.ComputeClient
	vcnClient     core.VirtualNetworkClient
	cfg           *Config
	context       context.Context
}

var retryPolicy = &common.RetryPolicy{
	MaximumNumberAttempts: 10,
	ShouldRetryOperation: func(res common.OCIOperationResponse) bool {
		var e common.ServiceError
		if errors.As(res.Error, &e) {
			switch e.GetHTTPStatusCode() {
			case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
				return true
			}
		}
		return false
	},
	NextDuration: func(res common.OCIOperationResponse) time.Duration {
		x := uint64(res.AttemptNumber)
		d := time.Duration(math.Pow(2, float64(atomic.LoadUint64(&x)))) * time.Second
		j := time.Duration(rand.Float64()*(2000)) * time.Millisecond
		w := d + j
		return w
	},
}

var requestMetadata = common.RequestMetadata{
	RetryPolicy: retryPolicy,
}

// NewDriverOCI Creates a new driverOCI with a connected compute client and a connected vcn client.
func NewDriverOCI(cfg *Config) (Driver, error) {
	coreClient, err := core.NewComputeClientWithConfigurationProvider(cfg.configProvider)
	if err != nil {
		return nil, err
	}

	vcnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.configProvider)
	if err != nil {
		return nil, err
	}

	return &driverOCI{
		computeClient: coreClient,
		vcnClient:     vcnClient,
		cfg:           cfg,
	}, nil
}

// CreateInstance creates a new compute instance.
func (d *driverOCI) CreateInstance(ctx context.Context, publicKey string) (string, error) {
	metadata := map[string]string{
		"ssh_authorized_keys": publicKey,
	}
	if d.cfg.Metadata != nil {
		for key, value := range d.cfg.Metadata {
			metadata[key] = value
		}
	}
	if d.cfg.UserData != "" {
		metadata["user_data"] = d.cfg.UserData
	}

	// Create VNIC details for instance
	CreateVnicDetails := core.CreateVnicDetails{
		AssignPublicIp:      d.cfg.CreateVnicDetails.AssignPublicIp,
		DisplayName:         d.cfg.CreateVnicDetails.DisplayName,
		HostnameLabel:       d.cfg.CreateVnicDetails.HostnameLabel,
		NsgIds:              d.cfg.CreateVnicDetails.NsgIds,
		PrivateIp:           d.cfg.CreateVnicDetails.PrivateIp,
		SkipSourceDestCheck: d.cfg.CreateVnicDetails.SkipSourceDestCheck,
		SubnetId:            d.cfg.CreateVnicDetails.SubnetId,
		DefinedTags:         d.cfg.CreateVnicDetails.DefinedTags,
		FreeformTags:        d.cfg.CreateVnicDetails.FreeformTags,
	}

	// Determine base image ID
	var imageId *string
	if d.cfg.BaseImageID != "" {
		imageId = &d.cfg.BaseImageID
	} else {
		// Pull images and determine which image ID to use, if BaseImageId not specified
		response, err := d.computeClient.ListImages(ctx, core.ListImagesRequest{
			CompartmentId:          d.cfg.BaseImageFilter.CompartmentId,
			DisplayName:            d.cfg.BaseImageFilter.DisplayName,
			OperatingSystem:        d.cfg.BaseImageFilter.OperatingSystem,
			OperatingSystemVersion: d.cfg.BaseImageFilter.OperatingSystemVersion,
			Shape:                  d.cfg.BaseImageFilter.Shape,
			LifecycleState:         "AVAILABLE",
			SortBy:                 "TIMECREATED",
			SortOrder:              "DESC",
			RequestMetadata:        requestMetadata,
		})
		if err != nil {
			return "", err
		}
		if len(response.Items) == 0 {
			return "", errors.New("base_image_filter returned no images")
		}
		if d.cfg.BaseImageFilter.DisplayNameSearch != nil {
			// Return most recent image that matches regex
			imageNameRegex, err := regexp.Compile(*d.cfg.BaseImageFilter.DisplayNameSearch)
			if err != nil {
				return "", err
			}
			for _, image := range response.Items {
				if imageNameRegex.MatchString(*image.DisplayName) {
					imageId = image.Id
					break
				}
			}
			if imageId == nil {
				return "", errors.New("No image matched display_name_search criteria")
			}
		} else {
			// If no regex provided, simply return most recent image pulled
			imageId = response.Items[0].Id
		}
	}

	// Create Source details which will be used to Launch Instance
	InstanceSourceDetails := core.InstanceSourceViaImageDetails{
		ImageId:             imageId,
		BootVolumeSizeInGBs: &d.cfg.BootVolumeSizeInGBs,
	}

	// Build instance details
	instanceDetails := core.LaunchInstanceDetails{
		AvailabilityDomain: &d.cfg.AvailabilityDomain,
		CompartmentId:      &d.cfg.CompartmentID,
		CreateVnicDetails:  &CreateVnicDetails,
		DefinedTags:        d.cfg.InstanceDefinedTags,
		DisplayName:        d.cfg.InstanceName,
		FreeformTags:       d.cfg.InstanceTags,
		Shape:              &d.cfg.Shape,
		SourceDetails:      InstanceSourceDetails,
		Metadata:           metadata,
	}

	instance, err := d.computeClient.LaunchInstance(context.TODO(), core.LaunchInstanceRequest{
		LaunchInstanceDetails: instanceDetails,
		RequestMetadata:       requestMetadata,
	})

	if err != nil {
		return "", err
	}

	return *instance.Id, nil
}

// CreateImage creates a new custom image.
func (d *driverOCI) CreateImage(ctx context.Context, id string) (core.Image, error) {
	res, err := d.computeClient.CreateImage(ctx, core.CreateImageRequest{CreateImageDetails: core.CreateImageDetails{
		CompartmentId: &d.cfg.ImageCompartmentID,
		InstanceId:    &id,
		DisplayName:   &d.cfg.ImageName,
		FreeformTags:  d.cfg.Tags,
		DefinedTags:   d.cfg.DefinedTags,
		LaunchMode:    core.CreateImageDetailsLaunchModeEnum(d.cfg.LaunchMode),
	},
		RequestMetadata: requestMetadata,
	})

	if err != nil {
		return core.Image{}, err
	}

	return res.Image, nil
}

// DeleteImage deletes a custom image.
func (d *driverOCI) DeleteImage(ctx context.Context, id string) error {
	_, err := d.computeClient.DeleteImage(ctx, core.DeleteImageRequest{
		ImageId:         &id,
		RequestMetadata: requestMetadata,
	})
	return err
}

// GetInstanceIP returns the public or private IP corresponding to the given instance id.
func (d *driverOCI) GetInstanceIP(ctx context.Context, id string) (string, error) {
	vnics, err := d.computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		InstanceId:      &id,
		CompartmentId:   &d.cfg.CompartmentID,
		RequestMetadata: requestMetadata,
	})
	if err != nil {
		return "", err
	}

	if len(vnics.Items) == 0 {
		return "", errors.New("instance has zero VNICs")
	}

	vnic, err := d.vcnClient.GetVnic(ctx, core.GetVnicRequest{
		VnicId:          vnics.Items[0].VnicId,
		RequestMetadata: requestMetadata,
	})
	if err != nil {
		return "", fmt.Errorf("Error getting VNIC details: %s", err)
	}

	if d.cfg.UsePrivateIP {
		return *vnic.PrivateIp, nil
	}

	if vnic.PublicIp == nil {
		return "", fmt.Errorf("Error getting VNIC Public Ip for: %s", id)
	}

	return *vnic.PublicIp, nil
}

func (d *driverOCI) GetInstanceInitialCredentials(ctx context.Context, id string) (string, string, error) {
	credentials, err := d.computeClient.GetWindowsInstanceInitialCredentials(ctx, core.GetWindowsInstanceInitialCredentialsRequest{
		InstanceId:      &id,
		RequestMetadata: requestMetadata,
	})
	if err != nil {
		return "", "", err
	}

	return *credentials.InstanceCredentials.Username, *credentials.InstanceCredentials.Password, err
}

// TerminateInstance terminates a compute instance.
func (d *driverOCI) TerminateInstance(ctx context.Context, id string) error {
	_, err := d.computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{
		InstanceId:      &id,
		RequestMetadata: requestMetadata,
	})
	return err
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverOCI) WaitForImageCreation(ctx context.Context, id string) error {
	return waitForResourceToReachState(
		func(string) (string, error) {
			image, err := d.computeClient.GetImage(ctx, core.GetImageRequest{
				ImageId:         &id,
				RequestMetadata: requestMetadata,
			})
			if err != nil {
				return "", err
			}
			return string(image.LifecycleState), nil
		},
		id,
		[]string{"PROVISIONING"},
		"AVAILABLE",
		0,             //Unlimited Retries
		5*time.Second, //5 second wait between retries
	)
}

// WaitForInstanceState waits for an instance to reach the a given terminal
// state.
func (d *driverOCI) WaitForInstanceState(ctx context.Context, id string, waitStates []string, terminalState string) error {
	return waitForResourceToReachState(
		func(string) (string, error) {
			instance, err := d.computeClient.GetInstance(ctx, core.GetInstanceRequest{
				InstanceId:      &id,
				RequestMetadata: requestMetadata,
			})
			if err != nil {
				return "", err
			}
			return string(instance.LifecycleState), nil
		},
		id,
		waitStates,
		terminalState,
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
