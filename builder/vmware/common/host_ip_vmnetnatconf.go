package common

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// VMnetNatConfIPFinder finds the IP address of the host machine by
// retrieving the IP from the vmnetnat.conf. This isn't a full proof
// technique but so far it has not failed.
type VMnetNatConfIPFinder struct{}

func (*VMnetNatConfIPFinder) HostIP() (string, error) {
	driver := &Workstation9Driver{}

	vmnetnat := driver.VmnetnatConfPath()
	if vmnetnat == "" {
		return "", errors.New("Could not find NAT vmnet conf file")
	}

	if _, err := os.Stat(vmnetnat); err != nil {
		return "", fmt.Errorf("Could not find NAT vmnet conf file: %s", vmnetnat)
	}

	f, err := os.Open(vmnetnat)
	if err != nil {
		return "", err
	}
	defer f.Close()

	ipRe := regexp.MustCompile(`^\s*ip\s*=\s*(.+?)\s*$`)

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			matches := ipRe.FindStringSubmatch(line)
			if matches != nil {
				ip := matches[1]
				dotIndex := strings.LastIndex(ip, ".")
				if dotIndex == -1 {
					continue
				}

				ip = ip[0:dotIndex] + ".1"
				return ip, nil
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}
	}

	return "", errors.New("host IP not found in " + vmnetnat)
}
