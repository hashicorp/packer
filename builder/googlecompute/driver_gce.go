package googlecompute

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"code.google.com/p/google-api-go-client/compute/v1"
	"github.com/mitchellh/packer/packer"

	// oauth2 "github.com/rasa/oauth2-fork-b3f9a68"
	"github.com/rasa/oauth2-fork-b3f9a68"

	// oauth2 "github.com/rasa/oauth2-fork-b3f9a68/google"
	"github.com/rasa/oauth2-fork-b3f9a68/google"
)

// driverGCE is a Driver implementation that actually talks to GCE.
// Create an instance using NewDriverGCE.
type driverGCE struct {
	projectId string
	service   *compute.Service
	ui        packer.Ui
}

var DriverScopes = []string{"https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/devstorage.full_control"}

func NewDriverGCE(ui packer.Ui, p string, a *accountFile) (Driver, error) {
	var f *oauth2.Options
	var err error

	// Auth with AccountFile first if provided
	if a.PrivateKey != "" {
		log.Printf("[INFO] Requesting Google token via AccountFile...")
		log.Printf("[INFO]   -- Email: %s", a.ClientEmail)
		log.Printf("[INFO]   -- Scopes: %s", DriverScopes)
		log.Printf("[INFO]   -- Private Key Length: %d", len(a.PrivateKey))

		f, err = oauth2.New(
			oauth2.JWTClient(a.ClientEmail, []byte(a.PrivateKey)),
			oauth2.Scope(DriverScopes...),
			google.JWTEndpoint())
	} else {
		log.Printf("[INFO] Requesting Google token via GCE Service Role...")

		f, err = oauth2.New(google.ComputeEngineAccount(""))
	}

	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating GCE client using...")
	service, err := compute.New(&http.Client{Transport: f.NewTransport()})
	if err != nil {
		return nil, err
	}

	return &driverGCE{
		projectId: p,
		service:   service,
		ui:        ui,
	}, nil
}

func (d *driverGCE) ImageExists(name string) bool {
	_, err := d.service.Images.Get(d.projectId, name).Do()
	// The API may return an error for reasons other than the image not
	// existing, but this heuristic is sufficient for now.
	return err == nil
}

func (d *driverGCE) CreateImage(name, description, zone, disk string) <-chan error {
	image := &compute.Image{
		Description: description,
		Name:        name,
		SourceDisk:  fmt.Sprintf("%s%s/zones/%s/disks/%s", d.service.BasePath, d.projectId, zone, disk),
		SourceType:  "RAW",
	}

	errCh := make(chan error, 1)
	op, err := d.service.Images.Insert(d.projectId, image).Do()
	if err != nil {
		errCh <- err
	} else {
		go waitForState(errCh, "DONE", d.refreshGlobalOp(op))
	}

	return errCh
}

func (d *driverGCE) DeleteImage(name string) <-chan error {
	errCh := make(chan error, 1)
	op, err := d.service.Images.Delete(d.projectId, name).Do()
	if err != nil {
		errCh <- err
	} else {
		go waitForState(errCh, "DONE", d.refreshGlobalOp(op))
	}

	return errCh
}

func (d *driverGCE) DeleteInstance(zone, name string) (<-chan error, error) {
	op, err := d.service.Instances.Delete(d.projectId, zone, name).Do()
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go waitForState(errCh, "DONE", d.refreshZoneOp(zone, op))
	return errCh, nil
}

func (d *driverGCE) DeleteDisk(zone, name string) (<-chan error, error) {
	op, err := d.service.Disks.Delete(d.projectId, zone, name).Do()
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go waitForState(errCh, "DONE", d.refreshZoneOp(zone, op))
	return errCh, nil
}

func (d *driverGCE) GetNatIP(zone, name string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, ni := range instance.NetworkInterfaces {
		if ni.AccessConfigs == nil {
			continue
		}

		for _, ac := range ni.AccessConfigs {
			if ac.NatIP != "" {
				return ac.NatIP, nil
			}
		}
	}

	return "", nil
}

func (d *driverGCE) RunInstance(c *InstanceConfig) (<-chan error, error) {
	// Get the zone
	d.ui.Message(fmt.Sprintf("Loading zone: %s", c.Zone))
	zone, err := d.service.Zones.Get(d.projectId, c.Zone).Do()
	if err != nil {
		return nil, err
	}

	// Get the image
	d.ui.Message(fmt.Sprintf("Loading image: %s in project %s", c.Image.Name, c.Image.ProjectId))
	image, err := d.getImage(c.Image)
	if err != nil {
		return nil, err
	}

	// Get the machine type
	d.ui.Message(fmt.Sprintf("Loading machine type: %s", c.MachineType))
	machineType, err := d.service.MachineTypes.Get(
		d.projectId, zone.Name, c.MachineType).Do()
	if err != nil {
		return nil, err
	}
	// TODO(mitchellh): deprecation warnings

	// Get the network
	d.ui.Message(fmt.Sprintf("Loading network: %s", c.Network))
	network, err := d.service.Networks.Get(d.projectId, c.Network).Do()
	if err != nil {
		return nil, err
	}

	// Build up the metadata
	metadata := make([]*compute.MetadataItems, len(c.Metadata))
	for k, v := range c.Metadata {
		metadata = append(metadata, &compute.MetadataItems{
			Key:   k,
			Value: v,
		})
	}

	// Create the instance information
	instance := compute.Instance{
		Description: c.Description,
		Disks: []*compute.AttachedDisk{
			&compute.AttachedDisk{
				Type:       "PERSISTENT",
				Mode:       "READ_WRITE",
				Kind:       "compute#attachedDisk",
				Boot:       true,
				AutoDelete: false,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: image.SelfLink,
					DiskSizeGb:  c.DiskSizeGb,
				},
			},
		},
		MachineType: machineType.SelfLink,
		Metadata: &compute.Metadata{
			Items: metadata,
		},
		Name: c.Name,
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Name: "AccessConfig created by Packer",
						Type: "ONE_TO_ONE_NAT",
					},
				},
				Network: network.SelfLink,
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			&compute.ServiceAccount{
				Email: "default",
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/compute",
					"https://www.googleapis.com/auth/devstorage.full_control",
				},
			},
		},
		Tags: &compute.Tags{
			Items: c.Tags,
		},
	}

	d.ui.Message("Requesting instance creation...")
	op, err := d.service.Instances.Insert(d.projectId, zone.Name, &instance).Do()
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go waitForState(errCh, "DONE", d.refreshZoneOp(zone.Name, op))
	return errCh, nil
}

func (d *driverGCE) WaitForInstance(state, zone, name string) <-chan error {
	errCh := make(chan error, 1)
	go waitForState(errCh, state, d.refreshInstanceState(zone, name))
	return errCh
}

func (d *driverGCE) getImage(img Image) (image *compute.Image, err error) {
	projects := []string{img.ProjectId, "centos-cloud", "coreos-cloud", "debian-cloud", "google-containers", "opensuse-cloud", "rhel-cloud", "suse-cloud", "ubuntu-os-cloud", "windows-cloud"}
	for _, project := range projects {
		image, err = d.service.Images.Get(project, img.Name).Do()
		if err == nil && image != nil && image.SelfLink != "" {
			return
		}
		image = nil
	}

	err = fmt.Errorf("Image %s could not be found in any of these projects: %s", img.Name, projects)
	return
}

func (d *driverGCE) refreshInstanceState(zone, name string) stateRefreshFunc {
	return func() (string, error) {
		instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
		if err != nil {
			return "", err
		}
		return instance.Status, nil
	}
}

func (d *driverGCE) refreshGlobalOp(op *compute.Operation) stateRefreshFunc {
	return func() (string, error) {
		newOp, err := d.service.GlobalOperations.Get(d.projectId, op.Name).Do()
		if err != nil {
			return "", err
		}

		// If the op is done, check for errors
		err = nil
		if newOp.Status == "DONE" {
			if newOp.Error != nil {
				for _, e := range newOp.Error.Errors {
					err = packer.MultiErrorAppend(err, fmt.Errorf(e.Message))
				}
			}
		}

		return newOp.Status, err
	}
}

func (d *driverGCE) refreshZoneOp(zone string, op *compute.Operation) stateRefreshFunc {
	return func() (string, error) {
		newOp, err := d.service.ZoneOperations.Get(d.projectId, zone, op.Name).Do()
		if err != nil {
			return "", err
		}

		// If the op is done, check for errors
		err = nil
		if newOp.Status == "DONE" {
			if newOp.Error != nil {
				for _, e := range newOp.Error.Errors {
					err = packer.MultiErrorAppend(err, fmt.Errorf(e.Message))
				}
			}
		}

		return newOp.Status, err
	}
}

// stateRefreshFunc is used to refresh the state of a thing and is
// used in conjunction with waitForState.
type stateRefreshFunc func() (string, error)

// waitForState will spin in a loop forever waiting for state to
// reach a certain target.
func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) {
	for {
		state, err := refresh()
		if err != nil {
			errCh <- err
			return
		}
		if state == target {
			errCh <- nil
			return
		}

		time.Sleep(2 * time.Second)
	}
}
