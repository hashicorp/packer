package common

import (
	"errors"
	"fmt"
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
		if matches != nil && strings.EqualFold(matches[1], f.MACAddress) && curLeaseEnd.Before(lastLeaseEnd) {
			curIp = lastIp
			curLeaseEnd = lastLeaseEnd
		}
	}

	if curIp == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases at %s", f.MACAddress, dhcpLeasesPath)
	}

	return curIp, nil
}
