//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type SharedImageGalleryDestination

package chroot

import (
	"fmt"
	"regexp"
)

// SharedImageGalleryDestination models an image version in a SIG
// that can be used as an source or destination for builders
type SharedImageGalleryDestination struct {
	ResourceGroup string `mapstructure:"resource_group"`
	GalleryName   string `mapstructure:"gallery_name"`
	ImageName     string `mapstructure:"image_name"`
	ImageVersion  string `mapstructure:"image_version"`

	TargetRegions     []TargetRegion `mapstructure:"target_regions"`
	ExcludeFromLatest bool           `mapstructure:"exlude_from_latest"`
}

// TargetRegion describes a region where the shared image should be replicated
type TargetRegion struct {
	// Name of the region
	Name string `mapstructure:"name"`
	// Number of replicas in this region. Default: 1
	ReplicaCount int32 `mapstructure:"replicas"`
	// Storage account type: Standard_LRS or Standard_ZRS. Default: Standard_ZRS
	StorageAccountType string `mapstructure:"storage_account_type"`
}

// ResourceID returns the resource ID string
func (sigd SharedImageGalleryDestination) ResourceID(subscriptionID string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/galleries/%s/images/%s/versions/%s",
		subscriptionID,
		sigd.ResourceGroup,
		sigd.GalleryName,
		sigd.ImageName,
		sigd.ImageVersion)
}

// Validate validates that the values in the shared image are valid (without checking them on the network)
func (sigd *SharedImageGalleryDestination) Validate(prefix string) (errs []error, warns []string) {
	if sigd.ResourceGroup == "" {
		errs = append(errs, fmt.Errorf("%s.resource_group is required", prefix))
	}
	if sigd.GalleryName == "" {
		errs = append(errs, fmt.Errorf("%s.gallery_name is required", prefix))
	}
	if sigd.ImageName == "" {
		errs = append(errs, fmt.Errorf("%s.image_name is required", prefix))
	}
	if match, err := regexp.MatchString("^[0-9]+\\.[0-9]+\\.[0-9]+$", sigd.ImageVersion); !match {
		if err != nil {
			warns = append(warns, fmt.Sprintf("Error matching pattern for %s.image_version: %s (this is probably a bug)", prefix, err))
		}
		errs = append(errs, fmt.Errorf("%s.image_version should match '^[0-9]+\\.[0-9]+\\.[0-9]+$'", prefix))
	}
	if len(sigd.TargetRegions) == 0 {
		warns = append(warns,
			fmt.Sprintf("%s.target_regions is empty; image will only be available in the region of the gallery", prefix))
	}
	return
}
