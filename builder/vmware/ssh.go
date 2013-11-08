package vmware

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
	"log"
	"os"
)

func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	vmxPath := state.Get("vmx_path").(string)

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

func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(*config)

	auth := []gossh.ClientAuth{
		gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
		gossh.ClientAuthKeyboardInteractive(
			ssh.PasswordKeyboardInteractive(config.SSHPassword)),
	}

	if config.SSHKeyPath != "" {
		keyring, err := sshKeyToKeyring(config.SSHKeyPath)
		if err != nil {
			return nil, err
		}

		auth = append(auth, gossh.ClientAuthKeyring(keyring))
	}

	return &gossh.ClientConfig{
		User: config.SSHUser,
		Auth: auth,
	}, nil
}

func sshKeyToKeyring(path string) (gossh.ClientKeyring, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	keyring := new(ssh.SimpleKeychain)
	if err := keyring.AddPEMKey(string(keyBytes)); err != nil {
		return nil, err
	}

	return keyring, nil
}
