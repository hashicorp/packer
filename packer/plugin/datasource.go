package plugin

import (
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

type cmdDataSource struct {
	d      packersdk.DataSource
	client *Client
}

func (d *cmdDataSource) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.ConfigSpec()
}

func (d *cmdDataSource) Configure(configs ...interface{}) error {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.Configure(configs...)
}

func (d *cmdDataSource) OutputSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.OutputSpec()
}

func (d *cmdDataSource) Execute() (cty.Value, error) {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.d.Execute()
}

func (d *cmdDataSource) checkExit(p interface{}, cb func()) {
	if d.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}
