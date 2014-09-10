// +build windows

package common

import (
	"fmt"
	"log"
	"regexp"
	"syscall"
)

func playerVerifyVersion(version string) error {
	key := `SOFTWARE\Wow6432Node\VMware, Inc.\VMware Player`
	subkey := "ProductVersion"
	productVersion, err := readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		key = `SOFTWARE\VMware, Inc.\VMware Player`
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
			`Could not find a VMware Player version in registry key %s\%s: '%s'`, key, subkey, productVersion)
	}
	log.Printf("Detected VMware Player version: %s", matches[1])

	return compareVersions(matches[1], version, "Player")
}
