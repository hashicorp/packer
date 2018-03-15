package common

import "fmt"

type VMConfig struct {
	VMName        string `mapstructure:"vm_name"`
	Folder        string `mapstructure:"folder"`
	Cluster       string `mapstructure:"cluster"`
	Host          string `mapstructure:"host"`
	ResourcePool  string `mapstructure:"resource_pool"`
	Datastore     string `mapstructure:"datastore"`
}

func (c *VMConfig) Prepare() []error {
	var errs []error

	if c.VMName == "" {
		errs = append(errs, fmt.Errorf("Target VM name is required"))
	}
	if c.Cluster == "" && c.Host == "" {
		errs = append(errs, fmt.Errorf("vSphere host or cluster is required"))
	}

	return errs
}
