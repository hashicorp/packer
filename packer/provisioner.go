package packer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/common"
)

// A provisioner is responsible for installing and configuring software
// on a machine prior to building the actual image.
type Provisioner interface {
	HCL2Speccer

	// Prepare is called with a set of configurations to setup the
	// internal state of the provisioner. The multiple configurations
	// should be merged in some sane way.
	Prepare(...interface{}) error

	// Provision is called to actually provision the machine. A context is
	// given for cancellation, a UI is given to communicate with the user, and
	// a communicator is given that is guaranteed to be connected to some
	// machine so that provisioning can be done.
	Provision(context.Context, Ui, Communicator, map[string]interface{}) error
}

// A HookedProvisioner represents a provisioner and information describing it
type HookedProvisioner struct {
	Provisioner Provisioner
	Config      interface{}
	TypeName    string
}

// A Hook implementation that runs the given provisioners.
type ProvisionHook struct {
	// The provisioners to run as part of the hook. These should already
	// be prepared (by calling Prepare) at some earlier stage.
	Provisioners []*HookedProvisioner
}

// Provisioners interpolate most of their fields in the prepare stage; this
// placeholder map helps keep fields that are only generated at build time from
// accidentally being interpolated into empty strings at prepare time.
// This helper function generates the most basic placeholder data which should
// be accessible to the provisioners. It is used to initialize provisioners, to
// force validation using the `generated` template function. In the future,
// custom generated data could be passed into provisioners from builders to
// enable specialized builder-specific (but still validated!!) access to builder
// data.
func BasicPlaceholderData() map[string]string {
	placeholderData := map[string]string{}
	msg := "Build_%s. " + common.PlaceholderMsg
	placeholderData["ID"] = fmt.Sprintf(msg, "ID")
	// The following correspond to communicator-agnostic functions that are
	// part of the SSH and WinRM communicator implementations. These functions
	// are not part of the communicator interface, but are stored on the
	// Communicator Config and return the appropriate values rather than
	// depending on the actual communicator config values. E.g "Password"
	// reprosents either WinRMPassword or SSHPassword, which makes this more
	// useful if a template contains multiple builds.
	placeholderData["Host"] = fmt.Sprintf(msg, "Host")
	placeholderData["Port"] = fmt.Sprintf(msg, "Port")
	placeholderData["User"] = fmt.Sprintf(msg, "User")
	placeholderData["Password"] = fmt.Sprintf(msg, "Password")
	placeholderData["ConnType"] = fmt.Sprintf(msg, "Type")
	placeholderData["PackerRunUUID"] = fmt.Sprintf(msg, "PackerRunUUID")
	placeholderData["SSHPublicKey"] = fmt.Sprintf(msg, "SSHPublicKey")
	placeholderData["SSHPrivateKey"] = fmt.Sprintf(msg, "SSHPrivateKey")

	// Backwards-compatability: WinRM Password can get through without forcing
	// the generated func validation.
	placeholderData["WinRMPassword"] = "{{.WinRMPassword}}"

	return placeholderData
}

func CastDataToMap(data interface{}) map[string]interface{} {
	// Provisioners expect a map[string]interface{} in their data field, but
	// it gets converted into a map[interface]interface on the way over the
	// RPC. Check that data can be cast into such a form, and cast it.
	cast := make(map[string]interface{})
	interMap, ok := data.(map[interface{}]interface{})
	if !ok {
		log.Printf("Unable to read map[string]interface out of data."+
			"Using empty interface: %#v", data)
	} else {
		for key, val := range interMap {
			keyString, ok := key.(string)
			if ok {
				cast[keyString] = val
			} else {
				log.Printf("Error casting generated data key to a string.")
			}
		}
	}
	return cast
}

// Runs the provisioners in order.
func (h *ProvisionHook) Run(ctx context.Context, name string, ui Ui, comm Communicator, data interface{}) error {
	// Shortcut
	if len(h.Provisioners) == 0 {
		return nil
	}

	if comm == nil {
		return fmt.Errorf(
			"No communicator found for provisioners! This is usually because the\n" +
				"`communicator` config was set to \"none\". If you have any provisioners\n" +
				"then a communicator is required. Please fix this to continue.")
	}
	for _, p := range h.Provisioners {
		ts := CheckpointReporter.AddSpan(p.TypeName, "provisioner", p.Config)

		cast := CastDataToMap(data)
		err := p.Provisioner.Provision(ctx, ui, comm, cast)

		ts.End(err)
		if err != nil {
			return err
		}
	}

	return nil
}

// PausedProvisioner is a Provisioner implementation that pauses before
// the provisioner is actually run.
type PausedProvisioner struct {
	PauseBefore time.Duration
	Provisioner Provisioner
}

func (p *PausedProvisioner) ConfigSpec() hcldec.ObjectSpec { return p.ConfigSpec() }
func (p *PausedProvisioner) FlatConfig() interface{}       { return p.FlatConfig() }
func (p *PausedProvisioner) Prepare(raws ...interface{}) error {
	return p.Provisioner.Prepare(raws...)
}

func (p *PausedProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator, generatedData map[string]interface{}) error {

	// Use a select to determine if we get cancelled during the wait
	ui.Say(fmt.Sprintf("Pausing %s before the next provisioner...", p.PauseBefore))
	select {
	case <-time.After(p.PauseBefore):
	case <-ctx.Done():
		return ctx.Err()
	}

	return p.Provisioner.Provision(ctx, ui, comm, generatedData)
}

// DebuggedProvisioner is a Provisioner implementation that waits until a key
// press before the provisioner is actually run.
type DebuggedProvisioner struct {
	Provisioner Provisioner

	cancelCh chan struct{}
	doneCh   chan struct{}
	lock     sync.Mutex
}

func (p *DebuggedProvisioner) ConfigSpec() hcldec.ObjectSpec { return p.ConfigSpec() }
func (p *DebuggedProvisioner) FlatConfig() interface{}       { return p.FlatConfig() }
func (p *DebuggedProvisioner) Prepare(raws ...interface{}) error {
	return p.Provisioner.Prepare(raws...)
}

func (p *DebuggedProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator, generatedData map[string]interface{}) error {
	// Use a select to determine if we get cancelled during the wait
	message := "Pausing before the next provisioner . Press enter to continue."

	result := make(chan string, 1)
	go func() {
		line, err := ui.Ask(message)
		if err != nil {
			log.Printf("Error asking for input: %s", err)
		}

		result <- line
	}()

	select {
	case <-result:
	case <-ctx.Done():
		return ctx.Err()
	}

	return p.Provisioner.Provision(ctx, ui, comm, generatedData)
}
