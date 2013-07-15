package vmware

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
	"log"
	"os"
)

func sshAddress(state map[string]interface{}) (string, error) {
	config := state["config"].(*config)
	vmxPath := state["vmx_path"].(string)

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
		if macAddress, ok = vmxData["ethernet0.generatedAddress"]; !ok || macAddress == "" {
			return "", errors.New("couldn't find MAC address in VMX")
		}
	}

	ipLookup := &DHCPLeaseGuestLookup{
		Device:     "vmnet8",
		MACAddress: macAddress,
	}

	ipAddress, err := ipLookup.GuestIP()
	if err != nil {
		log.Printf("IP lookup failed: %s", err)
		return "", fmt.Errorf("IP lookup failed: %s", err)
	}

	log.Printf("Detected IP: %s", ipAddress)
	return fmt.Sprintf("%s:%d", ipAddress, config.SSHPort), nil
}

func sshConfig(state map[string]interface{}) (*gossh.ClientConfig, error) {
	config := state["config"].(*config)

	return &gossh.ClientConfig{
		User: config.SSHUser,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
			gossh.ClientAuthKeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		},
	}, nil
}
