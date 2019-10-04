package chroot

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
)

func TestBuilder_Prepare(t *testing.T) {
	type config map[string]interface{}
	type regexMatchers map[string]string // map of regex : error message

	tests := []struct {
		name     string
		config   config
		validate func(Config)
		wantErr  bool
	}{
		{
			name: "HappyPathFromPlatformImage",
			config: config{
				"client_id":         "123",
				"client_secret":     "456",
				"subscription_id":   "789",
				"image_resource_id": "/subscriptions/789/resourceGroups/otherrgname/providers/Microsoft.Compute/images/MyDebianOSImage-{{timestamp}}",
				"source":            "credativ:Debian:9:latest",
			},
			validate: func(c Config) {
				if c.OSDiskSizeGB != 0 {
					t.Errorf("Expected OSDiskSizeGB to be 0, was %+v", c.OSDiskSizeGB)
				}
				if c.MountPartition != "1" {
					t.Errorf("Expected MountPartition to be %s, but found %s", "1", c.MountPartition)
				}
				if c.OSDiskStorageAccountType != string(compute.PremiumLRS) {
					t.Errorf("Expected OSDiskStorageAccountType to be %s, but found %s", string(compute.PremiumLRS), c.OSDiskStorageAccountType)
				}
				if c.OSDiskCacheType != string(compute.CachingTypesReadOnly) {
					t.Errorf("Expected OSDiskCacheType to be %s, but found %s", string(compute.CachingTypesReadOnly), c.OSDiskCacheType)
				}
				if c.ImageHyperVGeneration != string(compute.V1) {
					t.Errorf("Expected ImageHyperVGeneration to be %s, but found %s", string(compute.V1), c.ImageHyperVGeneration)
				}
			},
		},
		{
			name: "HappyPathFromPlatformImage",
			config: config{
				"image_resource_id": "/subscriptions/789/resourceGroups/otherrgname/providers/Microsoft.Compute/images/MyDebianOSImage-{{timestamp}}",
				"source":            "/subscriptions/789/resourceGroups/testrg/providers/Microsoft.Compute/disks/diskname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{}

			_, err := b.Prepare(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(b.config)
			}
		})
	}
}
