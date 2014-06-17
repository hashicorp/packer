package common

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
)

func SSHAddressFunc(config *SSHConfig) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		driver := state.Get("driver").(Driver)
		vmxPath := state.Get("vmx_path").(string)

		if config.SSHHost != "" {
			return fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort), nil
		}

		log.Println("Lookup up IP information...")
		f, err := os.Open(vmxPath)
		if err != nil {
			return "", err
		}
		defer f.Close()

		vmxBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}

		vmxData := ParseVMX(string(vmxBytes))

		var ok bool
		macAddress := ""
		if macAddress, ok = vmxData["ethernet0.address"]; !ok || macAddress == "" {
			if macAddress, ok = vmxData["ethernet0.generatedaddress"]; !ok || macAddress == "" {
				return "", errors.New("couldn't find MAC address in VMX")
			}
		}

		ipLookup := &DHCPLeaseGuestLookup{
			Driver:     driver,
			Device:     "vmnet8",
			MACAddress: macAddress,
		}

		ipAddress, err := ipLookup.GuestIP()
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			return "", fmt.Errorf("IP lookup failed: %s", err)
		}

		if ipAddress == "" {
			log.Println("IP is blank, no IP yet.")
			return "", errors.New("IP is blank")
		}

		log.Printf("Detected IP: %s", ipAddress)
		return fmt.Sprintf("%s:%d", ipAddress, config.SSHPort), nil
	}
}

func SSHConfigFunc(config *SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		auth := []gossh.AuthMethod{
			gossh.Password(config.SSHPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			signer, err := sshKeyToSigner(config.SSHKeyPath)
			if err != nil {
				return nil, err
			}

			auth = append(auth, gossh.PublicKeys(signer))
		}

		return &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: auth,
		}, nil
	}
}

func sshKeyToSigner(path string) (gossh.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	signer, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}
