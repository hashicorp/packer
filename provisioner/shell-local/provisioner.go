package shell

import (
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	config sl.Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	// This is a bit of a hack. For provisioners that need access to
	// auto-generated WinRMPasswords, the mechanism of keeping provisioner data
	// and build data totally segregated breaks down. We get around this by
	// having the builder stash the WinRMPassword in the state bag, then
	// grabbing it out of the statebag inside of StepProvision.  Then, when
	// the time comes to provision for real, we run the prepare step one more
	// time, now with WinRMPassword defined in the raws, and can store the
	// password on the provisioner config without overwriting the rest of the
	// work we've already done in the first prepare run.
	if len(raws) == 1 {
		for k, v := range raws[0].(map[interface{}]interface{}) {
			if k.(string) == "WinRMPassword" {
				p.config.WinRMPassword = v.(string)
				// Even if WinRMPassword is not gonna be used, we've stored the
				// key and pointed it to an empty string. That means we'll
				// always reach this on our second-run of Prepare()
				return nil
			}
		}
	}

	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}

	err = sl.Validate(&p.config)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, _ packer.Communicator) error {
	_, retErr := sl.Run(ui, &p.config)
	if retErr != nil {
		return retErr
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just do nothing. When the process ends, so will our provisioner
}
