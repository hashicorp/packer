//go:generate struct-markdown

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
	// The ID or full URL to the base image to use. This
    // is the image that will be used to launch a new server and provision it.
    // Unless you specify completely custom SSH settings, the source image must
    // have cloud-init installed so that the keypair gets assigned properly.
	SourceImage        string            `mapstructure:"source_image" required:"true"`
	// The name of the base image to use. This is
    // an alternative way of providing source_image and only either of them can
    // be specified.
	SourceImageName    string            `mapstructure:"source_image_name" required:"true"`
	// The search filters for determining the base
    // image to use. This is an alternative way of providing source_image and
    // only one of these methods can be used. source_image will override the
    // filters.
	SourceImageFilters ImageFilter       `mapstructure:"source_image_filter" required:"true"`
	// The ID, name, or full URL for the desired flavor for
    // the server to be created.
	Flavor             string            `mapstructure:"flavor" required:"true"`
	// The availability zone to launch the server
    // in. If this isn't specified, the default enforced by your OpenStack cluster
    // will be used. This may be required for some OpenStack clusters.
	AvailabilityZone   string            `mapstructure:"availability_zone" required:"false"`
	// For rackspace, whether or not to wait for
    // Rackconnect to assign the machine an IP address before connecting via SSH.
    // Defaults to false.
	RackconnectWait    bool              `mapstructure:"rackconnect_wait" required:"false"`
	// The ID or name of an external network that
    // can be used for creation of a new floating IP.
	FloatingIPNetwork  string            `mapstructure:"floating_ip_network" required:"false"`
	// A specific floating IP to assign to this instance.
	FloatingIP         string            `mapstructure:"floating_ip" required:"false"`
	// Whether or not to attempt to reuse existing
    // unassigned floating ips in the project before allocating a new one. Note
    // that it is not possible to safely do this concurrently, so if you are
    // running multiple openstack builds concurrently, or if other processes are
    // assigning and using floating IPs in the same openstack project while packer
    // is running, you should not set this to true. Defaults to false.
	ReuseIPs           bool              `mapstructure:"reuse_ips" required:"false"`
	// A list of security groups by name to
    // add to this instance.
	SecurityGroups     []string          `mapstructure:"security_groups" required:"false"`
	// A list of networks by UUID to attach to
    // this instance.
	Networks           []string          `mapstructure:"networks" required:"false"`
	// A list of ports by UUID to attach to this
    // instance.
	Ports              []string          `mapstructure:"ports" required:"false"`
	// User data to apply when launching the instance. Note
    // that you need to be careful about escaping characters due to the templates
    // being JSON. It is often more convenient to use user_data_file, instead.
    // Packer will not automatically wait for a user script to finish before
    // shutting down the instance this must be handled in a provisioner.
	UserData           string            `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
    // data when launching the instance.
	UserDataFile       string            `mapstructure:"user_data_file" required:"false"`
	// Name that is applied to the server instance
    // created by Packer. If this isn't specified, the default is same as
    // image_name.
	InstanceName       string            `mapstructure:"instance_name" required:"false"`
	// Metadata that is
    // applied to the server instance created by Packer. Also called server
    // properties in some documentation. The strings have a max size of 255 bytes
    // each.
	InstanceMetadata   map[string]string `mapstructure:"instance_metadata" required:"false"`
	// Whether to force the OpenStack instance to be
    // forcefully deleted. This is useful for environments that have
    // reclaim / soft deletion enabled. By default this is false.
	ForceDelete        bool              `mapstructure:"force_delete" required:"false"`
	// Whether or not nova should use ConfigDrive for
    // cloud-init metadata.
	ConfigDrive bool `mapstructure:"config_drive" required:"false"`
	// Deprecated use floating_ip_network
    // instead.
	FloatingIPPool string `mapstructure:"floating_ip_pool" required:"false"`
	// Use Block Storage service volume for
    // the instance root volume instead of Compute service local volume (default).
	UseBlockStorageVolume  bool   `mapstructure:"use_blockstorage_volume" required:"false"`
	// Name of the Block Storage service volume. If this
    // isn't specified, random string will be used.
	VolumeName             string `mapstructure:"volume_name" required:"false"`
	// Type of the Block Storage service volume. If this
    // isn't specified, the default enforced by your OpenStack cluster will be
    // used.
	VolumeType             string `mapstructure:"volume_type" required:"false"`
	// Size of the Block Storage service volume in GB. If
    // this isn't specified, it is set to source image min disk value (if set) or
    // calculated from the source image bytes size. Note that in some cases this
    // needs to be specified, if use_blockstorage_volume is true.
	VolumeSize             int    `mapstructure:"volume_size" required:"false"`
	// Availability zone of the Block
    // Storage service volume. If omitted, Compute instance availability zone will
    // be used. If both of Compute instance and Block Storage volume availability
    // zones aren't specified, the default enforced by your OpenStack cluster will
    // be used.
	VolumeAvailabilityZone string `mapstructure:"volume_availability_zone" required:"false"`

	// Not really used, but here for BC
	OpenstackProvider string `mapstructure:"openstack_provider"`
	// Deprecated use floating_ip or
    // floating_ip_pool instead.
	UseFloatingIp     bool   `mapstructure:"use_floating_ip" required:"false"`

	sourceImageOpts images.ListOpts
}

type ImageFilter struct {
	// filters used to select a source_image.
    // NOTE: This will fail unless exactly one image is returned, or
    // most_recent is set to true. Of the filters described in
    // ImageService, the
    // following are valid:
	Filters    ImageFilterOptions `mapstructure:"filters" required:"false"`
	// Selects the newest created image when true.
    // This is most useful for selecting a daily distro build.
	MostRecent bool               `mapstructure:"most_recent" required:"false"`
}

type ImageFilterOptions struct {
	Name       string            `mapstructure:"name"`
	Owner      string            `mapstructure:"owner"`
	Tags       []string          `mapstructure:"tags"`
	Visibility string            `mapstructure:"visibility"`
	Properties map[string]string `mapstructure:"properties"`
}

func (f *ImageFilterOptions) Empty() bool {
	return f.Name == "" && f.Owner == "" && len(f.Tags) == 0 && f.Visibility == "" && len(f.Properties) == 0
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
