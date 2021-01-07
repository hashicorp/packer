//go:generate mapstructure-to-hcl2 -type Config

package ami

import "github.com/hashicorp/packer-plugin-sdk/template/config"

type Config struct {
	config.KeyValueFilter `mapstructure:",squash"`
	Owners                []string
	MostRecent            bool `mapstructure:"most_recent"`
}

func (d *Config) GetOwners() []*string {
	res := make([]*string, 0, len(d.Owners))
	for _, owner := range d.Owners {
		i := owner
		res = append(res, &i)
	}
	return res
}

func (d *Config) Empty() bool {
	return len(d.Owners) == 0 && d.KeyValueFilter.Empty()
}

func (d *Config) NoOwner() bool {
	return len(d.Owners) == 0
}
