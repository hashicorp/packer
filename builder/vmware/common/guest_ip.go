package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// Interface to help find the IP address of a running virtual machine.
type GuestIPFinder interface {
	GuestIP() (string, error)
}

// DHCPLeaseGuestLookup looks up the IP address of a guest using DHCP
// lease information from the VMware network devices.
type DHCPLeaseGuestLookup struct {
	// Driver that is being used (to find leases path)
	Driver Driver

	// Device that the guest is connected to.
	Device string

	// MAC address of the guest.
	MACAddress string
}

func IPAddressFunc() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		return LookupGuestIPAddress(state)
	}
}

// Finds the Guest IP address from VMWare
func LookupGuestIPAddress(state multistep.StateBag) (string, error) {
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
	return ipAddress, nil
}

func (f *DHCPLeaseGuestLookup) GuestIP() (string, error) {
	dhcpLeasesPath := f.Driver.DhcpLeasesPath(f.Device)
	log.Printf("DHCP leases path: %s", dhcpLeasesPath)
	if dhcpLeasesPath == "" {
		return "", errors.New("no DHCP leases path found.")
	}

	fh, err := os.Open(dhcpLeasesPath)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	dhcpBytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return "", err
	}

	var lastIp string
	var lastLeaseEnd time.Time

	var curIp string
	var curLeaseEnd time.Time

	ipLineRe := regexp.MustCompile(`^lease (.+?) {$`)
	endTimeLineRe := regexp.MustCompile(`^\s*ends \d (.+?);$`)
	macLineRe := regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)

	for _, line := range strings.Split(string(dhcpBytes), "\n") {
		// Need to trim off CR character when running in windows
		line = strings.TrimRight(line, "\r")

		matches := ipLineRe.FindStringSubmatch(line)
		if matches != nil {
			lastIp = matches[1]
			continue
		}

		matches = endTimeLineRe.FindStringSubmatch(line)
		if matches != nil {
			lastLeaseEnd, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		// If the mac address matches and this lease ends farther in the
		// future than the last match we might have, then choose it.
		matches = macLineRe.FindStringSubmatch(line)
		if matches != nil && matches[1] == f.MACAddress && curLeaseEnd.Before(lastLeaseEnd) {
			curIp = lastIp
			curLeaseEnd = lastLeaseEnd
		}
	}

	if curIp == "" {
		return "", errors.New("IP not found for MAC in DHCP leases")
	}

	return curIp, nil
}
