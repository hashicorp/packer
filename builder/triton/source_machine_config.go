//go:generate struct-markdown

package triton

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// SourceMachineConfig represents the configuration to run a machine using
// the SDC API in order for provisioning to take place.
type SourceMachineConfig struct {
	// Name of the VM used for building the
    // image. Does not affect (and does not have to be the same) as the name for a
    // VM instance running this image. Maximum 512 characters but should in
    // practice be much shorter (think between 5 and 20 characters). For example
    // mysql-64-server-image-builder. When omitted defaults to
    // packer-builder-[image_name].
	MachineName            string             `mapstructure:"source_machine_name" required:"false"`
	// The Triton package to use while
    // building the image. Does not affect (and does not have to be the same) as
    // the package which will be used for a VM instance running this image. On the
    // Joyent public cloud this could for example be g3-standard-0.5-smartos.
	MachinePackage         string             `mapstructure:"source_machine_package" required:"true"`
	// The UUID of the image to base the new
    // image on. Triton supports multiple types of images, called 'brands' in
    // Triton / Joyent lingo, for contains and VM's. See the chapter Containers
    // and virtual machines in
    // the Joyent Triton documentation for detailed information. The following
    // brands are currently supported by this builder:joyent andkvm. The
    // choice of base image automatically decides the brand. On the Joyent public
    // cloud a valid source_machine_image could for example be
    // 70e3ae72-96b6-11e6-9056-9737fd4d0764 for version 16.3.1 of the 64bit
    // SmartOS base image (a 'joyent' brand image). source_machine_image_filter
    // can be used to populate this UUID.
	MachineImage           string             `mapstructure:"source_machine_image" required:"true"`
	// The UUID's of Triton
    // networks added to the source machine used for creating the image. For
    // example if any of the provisioners which are run need Internet access you
    // will need to add the UUID's of the appropriate networks here. If this is
    // not specified, instances will be placed into the default Triton public and
    // internal networks.
	MachineNetworks        []string           `mapstructure:"source_machine_networks" required:"false"`
	// Triton metadata
    // applied to the VM used to create the image. Metadata can be used to pass
    // configuration information to the VM without the need for networking. See
    // Using the metadata
    // API in the
    // Joyent documentation for more information. This can for example be used to
    // set the user-script metadata key to have Triton start a user supplied
    // script after the VM has booted.
	MachineMetadata        map[string]string  `mapstructure:"source_machine_metadata" required:"false"`
	// Tags applied to the
    // VM used to create the image.
	MachineTags            map[string]string  `mapstructure:"source_machine_tags" required:"false"`
	// Whether or not the firewall
    // of the VM used to create an image of is enabled. The Triton firewall only
    // filters inbound traffic to the VM. All outbound traffic is always allowed.
    // Currently this builder does not provide an interface to add specific
    // firewall rules. Unless you have a global rule defined in Triton which
    // allows SSH traffic enabling the firewall will interfere with the SSH
    // provisioner. The default is false.
	MachineFirewallEnabled bool               `mapstructure:"source_machine_firewall_enabled" required:"false"`
	// Filters used to populate the
    // source_machine_image field. Example:
	MachineImageFilters    MachineImageFilter `mapstructure:"source_machine_image_filter" required:"false"`
}

type MachineImageFilter struct {
	MostRecent bool `mapstructure:"most_recent"`
	Name       string
	OS         string
	Version    string
	Public     bool
	State      string
	Owner      string
	Type       string
}

func (m *MachineImageFilter) Empty() bool {
	return m.Name == "" && m.OS == "" && m.Version == "" && m.State == "" && m.Owner == "" && m.Type == ""
}

// Prepare performs basic validation on a SourceMachineConfig struct.
func (c *SourceMachineConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.MachinePackage == "" {
		errs = append(errs, fmt.Errorf("A source_machine_package must be specified"))
	}

	if c.MachineImage != "" && c.MachineImageFilters.Name != "" {
		errs = append(errs, fmt.Errorf("You cannot specify a Machine Image and also Machine Name filter"))
	}

	if c.MachineNetworks == nil {
		c.MachineNetworks = []string{}
	}

	if c.MachineMetadata == nil {
		c.MachineMetadata = make(map[string]string)
	}

	if c.MachineTags == nil {
		c.MachineTags = make(map[string]string)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
