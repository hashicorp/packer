package googlecompute

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCreateWindowsPassword represents a Packer build step that sets the windows password on a Windows GCE instance.
type StepCreateWindowsPassword struct {
	Debug        bool
	DebugKeyPath string
}

// Run executes the Packer build step that sets the windows password on a Windows GCE instance.
func (s *StepCreateWindowsPassword) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(Driver)
	c := state.Get("config").(*Config)
	name := state.Get("instance_name").(string)

	if c.Comm.WinRMPassword != "" {
		state.Put("winrm_password", c.Comm.WinRMPassword)
		packer.LogSecretFilter.Set(c.Comm.WinRMPassword)
		return multistep.ActionContinue
	}

	create, ok := state.GetOk("create_windows_password")

	if !ok || !create.(bool) {
		return multistep.ActionContinue

	}
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

	email := ""
	if c.account != nil {
		email = c.account.jwt.Email
	}

	data := WindowsPasswordConfig{
		key:      priv,
		UserName: c.Comm.WinRMUser,
		Modulus:  base64.StdEncoding.EncodeToString(priv.N.Bytes()),
		Exponent: base64.StdEncoding.EncodeToString(buf[1:]),
		Email:    email,
		ExpireOn: time.Now().Add(time.Minute * 5),
	}

	if s.Debug {

		priv_blk := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PrivateKey(priv),
		}

		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Write out the key
		err = pem.Encode(f, &priv_blk)
		f.Close()
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
	}

	errCh, err := d.CreateOrResetWindowsPassword(name, c.Zone, &data)

	if err == nil {
		ui.Message("Waiting for windows password to complete...")
		select {
		case err = <-errCh:
		case <-time.After(c.StateTimeout):
			err = errors.New("time out while waiting for the password to be created")
		}
	}

	if err != nil {
		err := fmt.Errorf("Error creating windows password: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Created password.")

	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled): %s", data.password))
	}

	state.Put("winrm_password", data.password)
	packer.LogSecretFilter.Set(data.password)

	return multistep.ActionContinue
}

// Nothing to clean up. The windows password is only created on the single instance.
func (s *StepCreateWindowsPassword) Cleanup(state multistep.StateBag) {}
