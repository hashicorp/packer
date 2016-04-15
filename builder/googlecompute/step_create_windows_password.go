package googlecompute

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateWindowsPassword represents a Packer build step that generates SSH key pairs.
type StepCreateWindowsPassword struct {
	Debug        bool
	DebugKeyPath string
}

// Run executes the Packer build step that generates SSH key pairs.
func (s *StepCreateWindowsPassword) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)
	config := state.Get("config").(*Config)
	name := state.Get("instance_name").(string)

	ui.Say("Creating windows user for instance...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err := fmt.Errorf("Error creating temporary key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	buf := make([]byte, 4)
    binary.BigEndian.PutUint32(buf, uint32(priv.E))
	
	data := WindowsPasswordConfig{
		key:      priv,
		UserName: config.Comm.WinRMUser,
		Modulus:  base64.StdEncoding.EncodeToString(priv.N.Bytes()),
		Exponent: base64.StdEncoding.EncodeToString(buf),
		Email:    config.account.ClientEmail,
		ExpireOn: time.Now().Add(time.Minute * 5),
	}
    
    ui.Message(fmt.Sprintf("%#v", data))

	errCh, err := driver.CreateOrResetWindowsPassword(name, config.Zone, &data)

	if err == nil {
		ui.Message("Waiting for windows password to complete...")
		select {
		case err = <-errCh:
		case <-time.After(config.stateTimeout):
			err = errors.New("time out while waiting for the password to be created")
		}
	}

	if err != nil {
		err := fmt.Errorf("Error creating windows password: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Created password %s", data.password))
	state.Put("windows_password", data.password)

	return multistep.ActionContinue
}

// Nothing to clean up. SSH keys are associated with a single GCE instance.
func (s *StepCreateWindowsPassword) Cleanup(state multistep.StateBag) {}
