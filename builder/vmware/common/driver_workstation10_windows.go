// +build windows

package common

import (
	"fmt"
	"log"
	"regexp"
	"syscall"
)

func workstationVerifyVersion(version string) error {
	key := `SOFTWARE\Wow6432Node\VMware, Inc.\VMware Workstation`
	subkey := "ProductVersion"
	productVersion, err := readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		key = `SOFTWARE\VMware, Inc.\VMware Workstation`
		productVersion, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
		if err != nil {
			log.Printf(`Unable to read registry key %s\%s`, key, subkey)
			return err
		}
	}

	versionRe := regexp.MustCompile(`^(\d+)\.`)
	matches := versionRe.FindStringSubmatch(productVersion)
	if matches == nil {
		return fmt.Errorf(
			`Could not find a VMware WS version in registry key %s\%s: '%s'`, key, subkey, productVersion)
	}
	log.Printf("Detected VMware WS version: %s", matches[1])

	return compareVersions(matches[1], version, "Workstation")
}
