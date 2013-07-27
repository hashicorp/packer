// +build windows
// Contributed by Ross Smith II (smithii.com)

package vmware

import (
	"errors"
	"regexp"
	"log"
	"strings"
	"net"
	"os"
	"io/ioutil"
)

// Interface to help find the host IP that is available from within
// the VMware virtual machines.
type HostIPFinder interface {
	HostIP() (string, error)
}

// IfconfigIPFinder finds the host IP based on the output of `ifconfig`.
type IfconfigIPFinder struct {
	Device string
}

func (f *IfconfigIPFinder) HostIP() (string, error) {
	ift, err := net.Interfaces()
	if err != nil {
		return "", errors.New("No network interfaces found")
	}

	vmwareMac, err := getVMWareMAC()
	if err != nil {
		log.Print(err)
	}

	log.Printf("Searching for MAC %s", vmwareMac)
	re := regexp.MustCompile("(?i)^" + vmwareMac)

	var rv string

	for _, ifi := range ift {
		log.Printf("Found interface %s", ifi.HardwareAddr.String())

		matches := re.FindStringSubmatch(ifi.HardwareAddr.String())

		if matches == nil {
			continue
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			log.Printf("No IP addresses found for %s", ifi.HardwareAddr.String())
			continue
		}

		for _, address := range addrs {
			rv = address.String()
			log.Printf("Found VMWare IP address %s", address.String())
		}

		// continue looping as VMNet8 comes after VMNet1 (at least on my system)
	}

	if rv > "" {
		return rv, nil
	}

	return "", errors.New("No VMWare MAC addresses found")
}

func getVMWareMAC() (string, error) {
	// return the first three tuples, if the actual MAC cannot be found
	const defaultMacRe = "00:50:56"

	programData := os.Getenv("ProgramData")
	vmnetnat := programData + "/VMware/vmnetnat.conf"
	if _, err := os.Stat(vmnetnat); os.IsNotExist(err) {
		log.Printf("File not found: '%s' (found '%s' in %%ProgramData%%)", vmnetnat, programData)
		return defaultMacRe, err
	}

	fh, err := os.Open(vmnetnat)
	if err != nil {
		return defaultMacRe, err
	}
	defer fh.Close()

	bytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return defaultMacRe, err
	}

	hostMacRe := regexp.MustCompile(`(?i)^\s*hostMAC\s*=\s*(.+)\s*$`)

	for _, line := range strings.Split(string(bytes), "\n") {
		// Need to trim off CR character when running in windows
		line = strings.TrimRight(line, "\r");

		matches := hostMacRe.FindStringSubmatch(line)
		if matches != nil {
			return matches[1], nil
		}
	}

	return defaultMacRe, nil

}
