package googlecompute

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/version"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/compute/v1"
	"strings"
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
	var err error

	var client *http.Client

	// Auth with AccountFile first if provided
	if a.PrivateKey != "" {
		log.Printf("[INFO] Requesting Google token via AccountFile...")
		log.Printf("[INFO]   -- Email: %s", a.ClientEmail)
		log.Printf("[INFO]   -- Scopes: %s", DriverScopes)
		log.Printf("[INFO]   -- Private Key Length: %d", len(a.PrivateKey))

		conf := jwt.Config{
			Email:      a.ClientEmail,
			PrivateKey: []byte(a.PrivateKey),
			Scopes:     DriverScopes,
			TokenURL:   "https://accounts.google.com/o/oauth2/token",
		}

		// Initiate an http.Client. The following GET request will be
		// authorized and authenticated on the behalf of
		// your service account.
		client = conf.Client(oauth2.NoContext)
	} else {
		log.Printf("[INFO] Requesting Google token via GCE Service Role...")
		client = &http.Client{
			Transport: &oauth2.Transport{
				// Fetch from Google Compute Engine's metadata server to retrieve
				// an access token for the provided account.
				// If no account is specified, "default" is used.
				Source: google.ComputeTokenSource(""),
			},
		}
	}

	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating GCE client...")
	service, err := compute.New(client)
	// Set UserAgent
	versionString := version.FormattedVersion()
	service.UserAgent = fmt.Sprintf(
		"(%s %s) Packer/%s", runtime.GOOS, runtime.GOARCH, versionString)

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

func (d *driverGCE) CreateImage(name, description, family, zone, disk string) (<-chan Image, <-chan error) {
	image := &compute.Image{
		Description: description,
		Name:        name,
		Family:      family,
		SourceDisk:  fmt.Sprintf("%s%s/zones/%s/disks/%s", d.service.BasePath, d.projectId, zone, disk),
		SourceType:  "RAW",
	}

	imageCh := make(chan Image, 1)
	errCh := make(chan error, 1)
	op, err := d.service.Images.Insert(d.projectId, image).Do()
	if err != nil {
		errCh <- err
	} else {
		go func() {
			err = waitForState(errCh, "DONE", d.refreshGlobalOp(op))
			if err != nil {
				close(imageCh)
			}
			image, err = d.getImage(name, d.projectId)
			if err != nil {
				close(imageCh)
				errCh <- err
			}
			imageCh <- Image{
				Name:      name,
				ProjectId: d.projectId,
				SizeGb:    image.DiskSizeGb,
			}
			close(imageCh)
		}()
	}

	return imageCh, errCh
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

func (d *driverGCE) GetInternalIP(zone, name string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, ni := range instance.NetworkInterfaces {
		if ni.NetworkIP == "" {
			continue
		}
		return ni.NetworkIP, nil
	}

	return "", nil
}

func (d *driverGCE) GetSerialPortOutput(zone, name string) (string, error) {
	output, err := d.service.Instances.GetSerialPortOutput(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}
	
	return output.Contents, nil
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
	image, err := d.getImage(c.Image.Name, c.Image.ProjectId)
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

	// Subnetwork
	// Validate Subnetwork config now that we have some info about the network
	if !network.AutoCreateSubnetworks && len(network.Subnetworks) > 0 {
		// Network appears to be in "custom" mode, so a subnetwork is required
		if c.Subnetwork == "" {
			return nil, fmt.Errorf("a subnetwork must be specified")
		}
	}
	// Get the subnetwork
	subnetworkSelfLink := ""
	if c.Subnetwork != "" {
		d.ui.Message(fmt.Sprintf("Loading subnetwork: %s for region: %s", c.Subnetwork, c.Region))
		subnetwork, err := d.service.Subnetworks.Get(d.projectId, c.Region, c.Subnetwork).Do()
		if err != nil {
			return nil, err
		}
		subnetworkSelfLink = subnetwork.SelfLink
	}

	var accessconfig *compute.AccessConfig
	// Use external IP if OmitExternalIP isn't set
	if !c.OmitExternalIP {
		accessconfig = &compute.AccessConfig{
			Name: "AccessConfig created by Packer",
			Type: "ONE_TO_ONE_NAT",
		}

		// If given a static IP, use it
		if c.Address != "" {
			region_url := strings.Split(zone.Region, "/")
			region := region_url[len(region_url)-1]
			address, err := d.service.Addresses.Get(d.projectId, region, c.Address).Do()
			if err != nil {
				return nil, err
			}
			accessconfig.NatIP = address.Address
		}
	}

	// Build up the metadata
	metadata := make([]*compute.MetadataItems, len(c.Metadata))
	for k, v := range c.Metadata {
		vCopy := v
		metadata = append(metadata, &compute.MetadataItems{
			Key:   k,
			Value: &vCopy,
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
					DiskType:    fmt.Sprintf("zones/%s/diskTypes/%s", zone.Name, c.DiskType),
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
				AccessConfigs: []*compute.AccessConfig{accessconfig},
				Network:       network.SelfLink,
				Subnetwork:    subnetworkSelfLink,
			},
		},
		Scheduling: &compute.Scheduling{
			Preemptible: c.Preemptible,
		},
		ServiceAccounts: []*compute.ServiceAccount{
			&compute.ServiceAccount{
				Email: c.ServiceAccountEmail,
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

func (d *driverGCE) getImage(name, projectId string) (image *compute.Image, err error) {
	projects := []string{projectId, "centos-cloud", "coreos-cloud", "debian-cloud", "google-containers", "opensuse-cloud", "rhel-cloud", "suse-cloud", "ubuntu-os-cloud", "windows-cloud"}
	for _, project := range projects {
		image, err = d.service.Images.Get(project, name).Do()
		if err == nil && image != nil && image.SelfLink != "" {
			return
		}
		image = nil
	}

	err = fmt.Errorf("Image %s could not be found in any of these projects: %s", name, projects)
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
func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) error {
	err := Retry(2, 2, 0, func() (bool, error) {
		state, err := refresh()
		if err != nil {
			return false, err
		} else if state == target {
			return true, nil
		}
		return false, nil
	})
	errCh <- err
	return err
}
