package googlecompute

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/useragent"
	"github.com/hashicorp/packer/packer"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// driverGCE is a Driver implementation that actually talks to GCE.
// Create an instance using NewDriverGCE.
type driverGCE struct {
	projectId string
	service   *compute.Service
	ui        packer.Ui
}

var DriverScopes = []string{"https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/devstorage.full_control"}

func NewClientGCE(conf *jwt.Config) (*http.Client, error) {
	var err error

	var client *http.Client

	// Auth with AccountFile first if provided
	if len(conf.PrivateKey) > 0 {
		log.Printf("[INFO] Requesting Google token via account_file...")
		log.Printf("[INFO]   -- Email: %s", conf.Email)
		log.Printf("[INFO]   -- Scopes: %s", DriverScopes)
		log.Printf("[INFO]   -- Private Key Length: %d", len(conf.PrivateKey))

		// Initiate an http.Client. The following GET request will be
		// authorized and authenticated on the behalf of
		// your service account.
		client = conf.Client(oauth2.NoContext)
	} else {
		log.Printf("[INFO] Requesting Google token via GCE API Default Client Token Source...")
		client, err = google.DefaultClient(oauth2.NoContext, DriverScopes...)
		// The DefaultClient uses the DefaultTokenSource of the google lib.
		// The DefaultTokenSource uses the "Application Default Credentials"
		// It looks for credentials in the following places, preferring the first location found:
		// 1. A JSON file whose path is specified by the
		//    GOOGLE_APPLICATION_CREDENTIALS environment variable.
		// 2. A JSON file in a location known to the gcloud command-line tool.
		//    On Windows, this is %APPDATA%/gcloud/application_default_credentials.json.
		//    On other systems, $HOME/.config/gcloud/application_default_credentials.json.
		// 3. On Google App Engine it uses the appengine.AccessToken function.
		// 4. On Google Compute Engine and Google App Engine Managed VMs, it fetches
		//    credentials from the metadata server.
		//    (In this final case any provided scopes are ignored.)
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewDriverGCE(ui packer.Ui, p string, conf *jwt.Config) (Driver, error) {
	client, err := NewClientGCE(conf)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Instantiating GCE client...")
	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	// Set UserAgent
	service.UserAgent = useragent.String()

	return &driverGCE{
		projectId: p,
		service:   service,
		ui:        ui,
	}, nil
}

func (d *driverGCE) CreateImage(name, description, family, zone, disk string, image_labels map[string]string, image_licenses []string, image_encryption_key *compute.CustomerEncryptionKey) (<-chan *Image, <-chan error) {
	gce_image := &compute.Image{
		Description:        description,
		Name:               name,
		Family:             family,
		Labels:             image_labels,
		Licenses:           image_licenses,
		ImageEncryptionKey: image_encryption_key,
		SourceDisk:         fmt.Sprintf("%s%s/zones/%s/disks/%s", d.service.BasePath, d.projectId, zone, disk),
		SourceType:         "RAW",
	}

	imageCh := make(chan *Image, 1)
	errCh := make(chan error, 1)
	op, err := d.service.Images.Insert(d.projectId, gce_image).Do()
	if err != nil {
		errCh <- err
	} else {
		go func() {
			err = waitForState(errCh, "DONE", d.refreshGlobalOp(op))
			if err != nil {
				close(imageCh)
				errCh <- err
				return
			}
			var image *Image
			image, err = d.GetImageFromProject(d.projectId, name, false)
			if err != nil {
				close(imageCh)
				errCh <- err
				return
			}
			imageCh <- image
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

func (d *driverGCE) GetImage(name string, fromFamily bool) (*Image, error) {
	projects := []string{
		d.projectId,
		// Public projects, drawn from
		// https://cloud.google.com/compute/docs/images
		"centos-cloud",
		"cos-cloud",
		"coreos-cloud",
		"debian-cloud",
		"rhel-cloud",
		"rhel-sap-cloud",
		"suse-cloud",
		"suse-sap-cloud",
		"ubuntu-os-cloud",
		"windows-cloud",
		"windows-sql-cloud",
		"gce-uefi-images",
		"gce-nvme",
		// misc
		"google-containers",
		"opensuse-cloud",
	}
	var errs error
	for _, project := range projects {
		image, err := d.GetImageFromProject(project, name, fromFamily)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		if image != nil {
			return image, nil
		}
	}

	return nil, fmt.Errorf(
		"Could not find image, %s, in projects, %s: %s", name,
		projects, errs)
}

func (d *driverGCE) GetImageFromProject(project, name string, fromFamily bool) (*Image, error) {
	var (
		image *compute.Image
		err   error
	)

	if fromFamily {
		image, err = d.service.Images.GetFromFamily(project, name).Do()
	} else {
		image, err = d.service.Images.Get(project, name).Do()
	}

	if err != nil {
		return nil, err
	} else if image == nil || image.SelfLink == "" {
		return nil, fmt.Errorf("Image, %s, could not be found in project: %s", name, project)
	} else {
		return &Image{
			Licenses:  image.Licenses,
			Name:      image.Name,
			ProjectId: project,
			SelfLink:  image.SelfLink,
			SizeGb:    image.DiskSizeGb,
		}, nil
	}
}

func (d *driverGCE) GetInstanceMetadata(zone, name, key string) (string, error) {
	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	if err != nil {
		return "", err
	}

	for _, item := range instance.Metadata.Items {
		if item.Key == key {
			return *item.Value, nil
		}
	}

	return "", fmt.Errorf("Instance metadata key, %s, not found.", key)
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

func (d *driverGCE) ImageExists(name string) bool {
	_, err := d.GetImageFromProject(d.projectId, name, false)
	// The API may return an error for reasons other than the image not
	// existing, but this heuristic is sufficient for now.
	return err == nil
}

func (d *driverGCE) RunInstance(c *InstanceConfig) (<-chan error, error) {
	// Get the zone
	d.ui.Message(fmt.Sprintf("Loading zone: %s", c.Zone))
	zone, err := d.service.Zones.Get(d.projectId, c.Zone).Do()
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

	networkId, subnetworkId, err := getNetworking(c)
	if err != nil {
		return nil, err
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

	var guestAccelerators []*compute.AcceleratorConfig
	if c.AcceleratorCount > 0 {
		ac := &compute.AcceleratorConfig{
			AcceleratorCount: c.AcceleratorCount,
			AcceleratorType:  c.AcceleratorType,
		}
		guestAccelerators = append(guestAccelerators, ac)
	}

	// Configure the instance's service account. If the user has set
	// disable_default_service_account, then the default service account
	// will not be used. If they also do not set service_account_email, then
	// the instance will be created with no service account or scopes.
	serviceAccount := &compute.ServiceAccount{}
	if !c.DisableDefaultServiceAccount {
		serviceAccount.Email = "default"
		serviceAccount.Scopes = c.Scopes
	}
	if c.ServiceAccountEmail != "" {
		serviceAccount.Email = c.ServiceAccountEmail
		serviceAccount.Scopes = c.Scopes
	}

	// Create the instance information
	instance := compute.Instance{
		Description: c.Description,
		Disks: []*compute.AttachedDisk{
			{
				Type:       "PERSISTENT",
				Mode:       "READ_WRITE",
				Kind:       "compute#attachedDisk",
				Boot:       true,
				AutoDelete: false,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: c.Image.SelfLink,
					DiskSizeGb:  c.DiskSizeGb,
					DiskType:    fmt.Sprintf("zones/%s/diskTypes/%s", zone.Name, c.DiskType),
				},
			},
		},
		GuestAccelerators: guestAccelerators,
		Labels:            c.Labels,
		MachineType:       machineType.SelfLink,
		Metadata: &compute.Metadata{
			Items: metadata,
		},
		MinCpuPlatform: c.MinCpuPlatform,
		Name:           c.Name,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{accessconfig},
				Network:       networkId,
				Subnetwork:    subnetworkId,
			},
		},
		Scheduling: &compute.Scheduling{
			OnHostMaintenance: c.OnHostMaintenance,
			Preemptible:       c.Preemptible,
		},
		ServiceAccounts: []*compute.ServiceAccount{
			serviceAccount,
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

func (d *driverGCE) CreateOrResetWindowsPassword(instance, zone string, c *WindowsPasswordConfig) (<-chan error, error) {

	errCh := make(chan error, 1)
	go d.createWindowsPassword(errCh, instance, zone, c)

	return errCh, nil
}

func (d *driverGCE) createWindowsPassword(errCh chan<- error, name, zone string, c *WindowsPasswordConfig) {

	data, err := json.Marshal(c)

	if err != nil {
		errCh <- err
		return
	}
	dCopy := string(data)

	instance, err := d.service.Instances.Get(d.projectId, zone, name).Do()
	instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{Key: "windows-keys", Value: &dCopy})

	op, err := d.service.Instances.SetMetadata(d.projectId, zone, name, &compute.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items:       instance.Metadata.Items,
	}).Do()

	if err != nil {
		errCh <- err
		return
	}

	newErrCh := make(chan error, 1)
	go waitForState(newErrCh, "DONE", d.refreshZoneOp(zone, op))

	select {
	case err = <-newErrCh:
	case <-time.After(time.Second * 30):
		err = errors.New("time out while waiting for instance to create")
	}

	if err != nil {
		errCh <- err
		return
	}

	timeout := time.Now().Add(time.Minute * 3)
	hash := sha1.New()
	random := rand.Reader

	for time.Now().Before(timeout) {
		if passwordResponses, err := d.getPasswordResponses(zone, name); err == nil {
			for _, response := range passwordResponses {
				if response.Modulus == c.Modulus {

					decodedPassword, err := base64.StdEncoding.DecodeString(response.EncryptedPassword)

					if err != nil {
						errCh <- err
						return
					}
					password, err := rsa.DecryptOAEP(hash, random, c.key, decodedPassword, nil)

					if err != nil {
						errCh <- err
						return
					}

					c.password = string(password)
					errCh <- nil
					return
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
	err = errors.New("Could not retrieve password. Timed out.")

	errCh <- err
	return

}

func (d *driverGCE) getPasswordResponses(zone, instance string) ([]windowsPasswordResponse, error) {
	output, err := d.service.Instances.GetSerialPortOutput(d.projectId, zone, instance).Port(4).Do()

	if err != nil {
		return nil, err
	}

	responses := strings.Split(output.Contents, "\n")

	passwordResponses := make([]windowsPasswordResponse, 0, len(responses))

	for _, response := range responses {
		var passwordResponse windowsPasswordResponse
		if err := json.Unmarshal([]byte(response), &passwordResponse); err != nil {
			continue
		}

		passwordResponses = append(passwordResponses, passwordResponse)
	}

	return passwordResponses, nil
}

func (d *driverGCE) WaitForInstance(state, zone, name string) <-chan error {
	errCh := make(chan error, 1)
	go waitForState(errCh, state, d.refreshInstanceState(zone, name))
	return errCh
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

// used in conjunction with waitForState.
type stateRefreshFunc func() (string, error)

// waitForState will spin in a loop forever waiting for state to
// reach a certain target.
func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) error {
	ctx := context.TODO()
	err := retry.Config{
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 2 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		state, err := refresh()
		if err != nil {
			return err
		}
		if state == target {
			return nil
		}
		return fmt.Errorf("retrying for state %s, got %s", target, state)
	})
	errCh <- err
	return err
}
