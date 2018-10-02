package openstack

import (
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// image and details on how to access that launched image.
type RunConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	SourceImage        string            `mapstructure:"source_image"`
	SourceImageName    string            `mapstructure:"source_image_name"`
	SourceImageFilters ImageFilter       `mapstructure:"source_image_filter"`
	Flavor             string            `mapstructure:"flavor"`
	AvailabilityZone   string            `mapstructure:"availability_zone"`
	RackconnectWait    bool              `mapstructure:"rackconnect_wait"`
	FloatingIPNetwork  string            `mapstructure:"floating_ip_network"`
	FloatingIP         string            `mapstructure:"floating_ip"`
	ReuseIPs           bool              `mapstructure:"reuse_ips"`
	SecurityGroups     []string          `mapstructure:"security_groups"`
	Networks           []string          `mapstructure:"networks"`
	Ports              []string          `mapstructure:"ports"`
	UserData           string            `mapstructure:"user_data"`
	UserDataFile       string            `mapstructure:"user_data_file"`
	InstanceName       string            `mapstructure:"instance_name"`
	InstanceMetadata   map[string]string `mapstructure:"instance_metadata"`

	ConfigDrive bool `mapstructure:"config_drive"`

	// Used for BC, value will be passed to the "floating_ip_network"
	FloatingIPPool string `mapstructure:"floating_ip_pool"`

	UseBlockStorageVolume  bool   `mapstructure:"use_blockstorage_volume"`
	VolumeName             string `mapstructure:"volume_name"`
	VolumeType             string `mapstructure:"volume_type"`
	VolumeAvailabilityZone string `mapstructure:"volume_availability_zone"`

	// Not really used, but here for BC
	OpenstackProvider string `mapstructure:"openstack_provider"`
	UseFloatingIp     bool   `mapstructure:"use_floating_ip"`

	sourceImageOpts images.ListOpts
}

type ImageFilter struct {
	Filters    ImageFilterOptions `mapstructure:"filters"`
	MostRecent bool               `mapstructure:"most_recent"`
}

type ImageFilterOptions struct {
	Name       string   `mapstructure:"name"`
	Owner      string   `mapstructure:"owner"`
	Tags       []string `mapstructure:"tags"`
	Visibility string   `mapstructure:"visibility"`
}

func (f *ImageFilterOptions) Empty() bool {
	return f.Name == "" && f.Owner == "" && len(f.Tags) == 0 && f.Visibility == ""
}

func (f *ImageFilterOptions) Build() (*images.ListOpts, error) {
	opts := images.ListOpts{}
	// Set defaults for status, member_status, and sort
	opts.Status = images.ImageStatusActive
	opts.MemberStatus = images.ImageMemberStatusAccepted
	opts.Sort = "created_at:desc"

	var err error

	if f.Name != "" {
		opts.Name = f.Name
	}
	if f.Owner != "" {
		opts.Owner = f.Owner
	}
	if len(f.Tags) > 0 {
		opts.Tags = f.Tags
	}
	if f.Visibility != "" {
		v, err := getImageVisibility(f.Visibility)
		if err == nil {
			opts.Visibility = *v
		}
	}

	return &opts, err
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided and we are not using
	// ssh_password.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.FloatingIPPool != "" && c.FloatingIPNetwork == "" {
		c.FloatingIPNetwork = c.FloatingIPPool
	}

	// Validation
	errs := c.Comm.Prepare(ctx)

	if c.Comm.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
			errs = append(errs, errors.New("A ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKeyFile == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, errors.New("A ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

	if c.SourceImage == "" && c.SourceImageName == "" && c.SourceImageFilters.Filters.Empty() {
		errs = append(errs, errors.New("Either a source_image, a source_image_name, or source_image_filter must be specified"))
	} else if len(c.SourceImage) > 0 && len(c.SourceImageName) > 0 {
		errs = append(errs, errors.New("Only a source_image or a source_image_name can be specified, not both."))
	}

	if c.Flavor == "" {
		errs = append(errs, errors.New("A flavor must be specified"))
	}

	if c.Comm.SSHIPVersion != "" && c.Comm.SSHIPVersion != "4" && c.Comm.SSHIPVersion != "6" {
		errs = append(errs, errors.New("SSH IP version must be either 4 or 6"))
	}

	for key, value := range c.InstanceMetadata {
		if len(key) > 255 {
			errs = append(errs, fmt.Errorf("Instance metadata key too long (max 255 bytes): %s", key))
		}
		if len(value) > 255 {
			errs = append(errs, fmt.Errorf("Instance metadata value too long (max 255 bytes): %s", value))
		}
	}

	if c.UseBlockStorageVolume {
		// Use Compute instance availability zone for the Block Storage volume if
		// it's not provided.
		if c.VolumeAvailabilityZone == "" {
			c.VolumeAvailabilityZone = c.AvailabilityZone
		}

		// Use random name for the Block Storage volume if it's not provided.
		if c.VolumeName == "" {
			c.VolumeName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
		}
	}

	// if neither ID or image name is provided outside the filter, build the filter
	if len(c.SourceImage) == 0 && len(c.SourceImageName) == 0 {

		listOpts, filterErr := c.SourceImageFilters.Filters.Build()

		if filterErr != nil {
			errs = append(errs, filterErr)
		}
		c.sourceImageOpts = *listOpts
	}

	return errs
}

// Retrieve the specific ImageVisibility using the exported const from images
func getImageVisibility(visibility string) (*images.ImageVisibility, error) {
	visibilities := [...]images.ImageVisibility{
		images.ImageVisibilityPublic,
		images.ImageVisibilityPrivate,
		images.ImageVisibilityCommunity,
		images.ImageVisibilityShared,
	}

	for _, v := range visibilities {
		if string(v) == visibility {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("Not a valid visibility: %s", visibility)
}
