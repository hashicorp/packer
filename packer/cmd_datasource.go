package packer

import (
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

type cmdDatasource struct {
	d      packersdk.Datasource
	client *PluginClient
}

func (d *cmdDatasource) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.ConfigSpec()
}

func (d *cmdDatasource) Configure(configs ...interface{}) error {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.Configure(configs...)
}

func (d *cmdDatasource) OutputSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.OutputSpec()
}

func (d *cmdDatasource) Execute() (cty.Value, error) {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.Execute()
}

func (d *cmdDatasource) checkExit(p interface{}, cb func()) {
	if d.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}
