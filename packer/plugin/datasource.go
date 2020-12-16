package plugin

import (
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

type cmdDataSource struct {
	p      packersdk.DataSource
	client *Client
}

func (d *cmdDataSource) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.p.ConfigSpec()
}

func (d *cmdDataSource) Configure(configs ...interface{}) error {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.p.Configure(configs...)
}

func (d *cmdDataSource) Execute() (cty.Value, error) {
	defer func() {
		r := recover()
		d.checkExit(r, nil)
	}()

	return d.p.Execute()
}

func (d *cmdDataSource) checkExit(p interface{}, cb func()) {
	if d.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}
