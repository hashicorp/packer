package triton

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/joyent/triton-go/compute"
	terrors "github.com/joyent/triton-go/errors"
)

type driverTriton struct {
	client *Client
	ui     packersdk.Ui
}

func NewDriverTriton(ui packersdk.Ui, config Config) (Driver, error) {
	client, err := config.AccessConfig.CreateTritonClient()
	if err != nil {
		return nil, err
	}

	return &driverTriton{
		client: client,
		ui:     ui,
	}, nil
}

func (d *driverTriton) GetImage(config Config) (string, error) {
	computeClient, _ := d.client.Compute()
	images, err := computeClient.Images().List(context.Background(), &compute.ListImagesInput{
		Name:    config.MachineImageFilters.Name,
		OS:      config.MachineImageFilters.OS,
		Version: config.MachineImageFilters.Version,
		Public:  config.MachineImageFilters.Public,
		Type:    config.MachineImageFilters.Type,
		State:   config.MachineImageFilters.State,
		Owner:   config.MachineImageFilters.Owner,
	})
	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", errors.New("No images found in your search. Please refine your search criteria")
	}

	if len(images) > 1 {
		if !config.MachineImageFilters.MostRecent {
			return "", errors.New("More than 1 machine image was found in your search. Please refine your search criteria")
		} else {
			return mostRecentImages(images).ID, nil
		}
	} else {
		return images[0].ID, nil
	}
}

func (d *driverTriton) CreateImageFromMachine(machineId string, config Config) (string, error) {
	computeClient, _ := d.client.Compute()
	image, err := computeClient.Images().CreateFromMachine(context.Background(), &compute.CreateImageFromMachineInput{
		MachineID:   machineId,
		Name:        config.ImageName,
		Version:     config.ImageVersion,
		Description: config.ImageDescription,
		HomePage:    config.ImageHomepage,
		EULA:        config.ImageEULA,
		ACL:         config.ImageACL,
		Tags:        config.ImageTags,
	})
	if err != nil {
		return "", err
	}

	return image.ID, err
}

func (d *driverTriton) CreateMachine(config Config) (string, error) {
	computeClient, _ := d.client.Compute()
	input := &compute.CreateInstanceInput{
		Package:         config.MachinePackage,
		Image:           config.MachineImage,
		Metadata:        config.MachineMetadata,
		Tags:            config.MachineTags,
		FirewallEnabled: config.MachineFirewallEnabled,
	}

	if config.MachineName == "" {
		// If not supplied generate a name for the source VM: "packer-builder-[image_name]".
		// The version is not used because it can contain characters invalid for a VM name.
		input.Name = "packer-builder-" + config.ImageName
	} else {
		input.Name = config.MachineName
	}

	if len(config.MachineNetworks) > 0 {
		input.Networks = config.MachineNetworks
	}

	machine, err := computeClient.Instances().Create(context.Background(), input)
	if err != nil {
		return "", err
	}

	return machine.ID, nil
}

func (d *driverTriton) DeleteImage(imageId string) error {
	computeClient, _ := d.client.Compute()
	return computeClient.Images().Delete(context.Background(), &compute.DeleteImageInput{
		ImageID: imageId,
	})
}

func (d *driverTriton) DeleteMachine(machineId string) error {
	computeClient, _ := d.client.Compute()
	return computeClient.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
		ID: machineId,
	})
}

func (d *driverTriton) GetMachineIP(machineId string) (string, error) {
	computeClient, _ := d.client.Compute()
	machine, err := computeClient.Instances().Get(context.Background(), &compute.GetInstanceInput{
		ID: machineId,
	})
	if err != nil {
		return "", err
	}

	return machine.PrimaryIP, nil
}

func (d *driverTriton) StopMachine(machineId string) error {
	computeClient, _ := d.client.Compute()
	return computeClient.Instances().Stop(context.Background(), &compute.StopInstanceInput{
		InstanceID: machineId,
	})
}

// waitForMachineState uses the supplied client to wait for the state of
// the machine with the given ID to reach the state described in state.
// If timeout is reached before the machine reaches the required state, an
// error is returned. If the machine reaches the target state within the
// timeout, nil is returned.
func (d *driverTriton) WaitForMachineState(machineId string, state string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			computeClient, _ := d.client.Compute()
			machine, err := computeClient.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: machineId,
			})
			if machine == nil {
				return false, err
			}
			return machine.State == state, err
		},
		3*time.Second,
		timeout,
	)
}

// waitForMachineDeletion uses the supplied client to wait for the machine
// with the given ID to be deleted. It is expected that the API call to delete
// the machine has already been issued at this point.
func (d *driverTriton) WaitForMachineDeletion(machineId string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			computeClient, _ := d.client.Compute()
			_, err := computeClient.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: machineId,
			})
			if err != nil {
				// Return true only when we receive a 410 (Gone) response.  A 404
				// indicates that the machine is being deleted whereas a 410 indicates
				// that this process has completed.
				if terrors.IsSpecificStatusCode(err, http.StatusGone) {
					return true, nil
				}
			}

			return false, err
		},
		3*time.Second,
		timeout,
	)
}

func (d *driverTriton) WaitForImageCreation(imageId string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			computeClient, _ := d.client.Compute()
			image, err := computeClient.Images().Get(context.Background(), &compute.GetImageInput{
				ImageID: imageId,
			})
			if image == nil {
				return false, err
			}
			return image.State == "active", err
		},
		3*time.Second,
		timeout,
	)
}

func waitFor(f func() (bool, error), every, timeout time.Duration) error {
	start := time.Now()

	for time.Since(start) <= timeout {
		stop, err := f()
		if err != nil {
			return err
		}

		if stop {
			return nil
		}

		time.Sleep(every)
	}

	return errors.New("Timed out while waiting for resource change")
}

func mostRecentImages(images []*compute.Image) *compute.Image {
	return sortImages(images)[0]
}

type imageSort []*compute.Image

func sortImages(images []*compute.Image) []*compute.Image {
	sortedImages := images
	sort.Sort(sort.Reverse(imageSort(sortedImages)))
	return sortedImages
}

func (a imageSort) Len() int {
	return len(a)
}

func (a imageSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a imageSort) Less(i, j int) bool {
	itime := a[i].PublishedAt
	jtime := a[j].PublishedAt
	return itime.Unix() < jtime.Unix()
}
