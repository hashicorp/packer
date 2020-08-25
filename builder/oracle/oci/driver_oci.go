package oci

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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

	instanceDetails := core.LaunchInstanceDetails{
		AvailabilityDomain: &d.cfg.AvailabilityDomain,
		CompartmentId:      &d.cfg.CompartmentID,
		DefinedTags:        d.cfg.InstanceDefinedTags,
		FreeformTags:       d.cfg.InstanceTags,
		ImageId:            &d.cfg.BaseImageID,
		Shape:              &d.cfg.Shape,
		SubnetId:           &d.cfg.SubnetID,
		Metadata:           metadata,
	}

	// When empty, the default display name is used.
	if d.cfg.InstanceName != "" {
		instanceDetails.DisplayName = &d.cfg.InstanceName
	}

	// Pass VNIC details, if specified, to the instance
	if len(d.cfg.CreateVnicDetails) > 0 {
		CreateVnicDetails, err := mapToCreateVnicDetails(d.cfg.CreateVnicDetails)
		if err != nil {
			return "", err
		}
		instanceDetails.CreateVnicDetails = &CreateVnicDetails
	}

	instance, err := d.computeClient.LaunchInstance(context.TODO(), core.LaunchInstanceRequest{LaunchInstanceDetails: instanceDetails})

	if err != nil {
		return "", err
	}

	return *instance.Id, nil
}

// CreateImage creates a new custom image.
func (d *driverOCI) CreateImage(ctx context.Context, id string) (core.Image, error) {
	res, err := d.computeClient.CreateImage(ctx, core.CreateImageRequest{CreateImageDetails: core.CreateImageDetails{
		CompartmentId: &d.cfg.CompartmentID,
		InstanceId:    &id,
		DisplayName:   &d.cfg.ImageName,
		FreeformTags:  d.cfg.Tags,
		DefinedTags:   d.cfg.DefinedTags,
	}})

	if err != nil {
		return core.Image{}, err
	}

	return res.Image, nil
}

// DeleteImage deletes a custom image.
func (d *driverOCI) DeleteImage(ctx context.Context, id string) error {
	_, err := d.computeClient.DeleteImage(ctx, core.DeleteImageRequest{ImageId: &id})
	return err
}

// GetInstanceIP returns the public or private IP corresponding to the given instance id.
func (d *driverOCI) GetInstanceIP(ctx context.Context, id string) (string, error) {
	vnics, err := d.computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		InstanceId:    &id,
		CompartmentId: &d.cfg.CompartmentID,
	})
	if err != nil {
		return "", err
	}

	if len(vnics.Items) == 0 {
		return "", errors.New("instance has zero VNICs")
	}

	vnic, err := d.vcnClient.GetVnic(ctx, core.GetVnicRequest{VnicId: vnics.Items[0].VnicId})
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
		InstanceId: &id,
	})
	if err != nil {
		return "", "", err
	}

	return *credentials.InstanceCredentials.Username, *credentials.InstanceCredentials.Password, err
}

// TerminateInstance terminates a compute instance.
func (d *driverOCI) TerminateInstance(ctx context.Context, id string) error {
	_, err := d.computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{
		InstanceId: &id,
	})
	return err
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverOCI) WaitForImageCreation(ctx context.Context, id string) error {
	return waitForResourceToReachState(
		func(string) (string, error) {
			image, err := d.computeClient.GetImage(ctx, core.GetImageRequest{ImageId: &id})
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
			instance, err := d.computeClient.GetInstance(ctx, core.GetInstanceRequest{InstanceId: &id})
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

// interfaceToBool converts a variable of type interface to type bool
func interfaceToBool(b interface{}) (bool, error) {
	var boolVal bool
	var err error

	switch t := b.(type) {
	case bool:
		boolVal = t
	case string:
		boolVal, err = strconv.ParseBool(t)
	default:
		boolVal, err = false, fmt.Errorf("failed to convert %v to boolean type", b)
	}

	return boolVal, err
}

// mapToCreateVnicDetails creates variable of type core.CreateVnicDetails from map
func mapToCreateVnicDetails(m map[string]interface{}) (core.CreateVnicDetails, error) {
	result := core.CreateVnicDetails{}

	if val, ok := m["assign_public_ip"]; ok {
		boolVal, err := interfaceToBool(val)
		if err != nil {
			return result, fmt.Errorf("assign_public_ip is incorrect type: %v", err)
		}
		result.AssignPublicIp = &boolVal
	}
	if val, ok := m["display_name"]; ok {
		tmp := val.(string)
		result.DisplayName = &tmp
	}
	if val, ok := m["hostname_label"]; ok {
		tmp := val.(string)
		result.HostnameLabel = &tmp
	}
	if val, ok := m["nsg_ids"]; ok {
		tmp, tmpok := val.([]interface{})
		if !tmpok {
			return result, errors.New("nsg_ids not in correct list format")
		}
		valStr := make([]string, len(tmp)) //convert []interface{} to []string
		for i, v := range tmp {
			valStr[i] = fmt.Sprint(v)
		}
		result.NsgIds = valStr
	}
	if val, ok := m["private_ip"]; ok {
		tmp := val.(string)
		result.PrivateIp = &tmp
	}
	if val, ok := m["skip_source_dest_check"]; ok {
		boolVal, err := interfaceToBool(val)
		if err != nil {
			return result, fmt.Errorf("skip_source_dest_check is incorrect type: %v", err)
		}
		result.SkipSourceDestCheck = &boolVal
	}
	if val, ok := m["subnet_id"]; ok {
		tmp := val.(string)
		result.SubnetId = &tmp
	}

	return result, nil
}
