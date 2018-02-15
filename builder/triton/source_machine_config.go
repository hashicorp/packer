package triton

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// SourceMachineConfig represents the configuration to run a machine using
// the SDC API in order for provisioning to take place.
type SourceMachineConfig struct {
	MachineName            string             `mapstructure:"source_machine_name"`
	MachinePackage         string             `mapstructure:"source_machine_package"`
	MachineImage           string             `mapstructure:"source_machine_image"`
	MachineNetworks        []string           `mapstructure:"source_machine_networks"`
	MachineMetadata        map[string]string  `mapstructure:"source_machine_metadata"`
	MachineTags            map[string]string  `mapstructure:"source_machine_tags"`
	MachineFirewallEnabled bool               `mapstructure:"source_machine_firewall_enabled"`
	MachineImageFilters    MachineImageFilter `mapstructure:"source_machine_image_filter"`
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
