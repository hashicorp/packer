package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
)

type stepCreateSSHKeyPair struct {
	keyName    string
	privateKey string
}

func (s *stepCreateSSHKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudstackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	// If we already have a private key for a pre loaded public
	// key on the base image we load that instead of creating a
	// SSH key pair.
	if c.SSHKeyPath != "" {
		ui.Say("Reading in SSH private key from local disk")

		f, err := os.Open(c.SSHKeyPath)
		if err != nil {
			return multistep.ActionHalt
		}
		defer f.Close()

		keyBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return multistep.ActionHalt
		}

		state.Put("ssh_private_key", string(keyBytes))
		state.Put("ssh_key_name", "")
		return multistep.ActionContinue
	}

	ui.Say("Creating temporary SSH key for virtual machine...")

	// The name of the public key on Cloudstack
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	response, err := client.CreateSSHKeyPair(name)
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.keyName = name
	s.privateKey = response.Createsshkeypairresponse.Keypair.Privatekey

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_name", name)
	state.Put("ssh_private_key", s.privateKey)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	client := state.Get("client").(*gopherstack.CloudstackClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary SSH key...")
	_, err := client.DeleteSSHKeyPair(s.keyName)
	if err != nil {
		log.Printf("Error cleaning up SSH key: %v", err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up SSH key. Please delete the key manually."))
	}
}
