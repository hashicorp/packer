package googlecompute

import (
	"strings"

	compute "google.golang.org/api/compute/v1"
)

type Image struct {
	GuestOsFeatures []*compute.GuestOsFeature
	Labels          map[string]string
	Licenses        []string
	Name            string
	ProjectId       string
	SelfLink        string
	SizeGb          int64
}

func (i *Image) IsWindows() bool {
	for _, license := range i.Licenses {
		if strings.Contains(license, "windows") {
			return true
		}
	}
	return false
}

func (i *Image) IsSecureBootCompatible() bool {
	for _, osFeature := range i.GuestOsFeatures {
		if osFeature.Type == "UEFI_COMPATIBLE" {
			return true
		}
	}
	return false
}
