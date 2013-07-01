package vmware

import (
	"errors"
	"io/ioutil"
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
	fh, err := os.Open(f.Driver.DhcpLeasesPath(f.Device))
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
