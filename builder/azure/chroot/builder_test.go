package chroot

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
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
			withMetadataStub(func() {
				b := &Builder{}

				_, _, err := b.Prepare(tt.config)

				if (err != nil) != tt.wantErr {
					t.Errorf("Builder.Prepare() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.validate != nil {
					tt.validate(b.config)
				}
			})
		})
	}
}

func Test_buildsteps(t *testing.T) {
	info := &client.ComputeInfo{
		Location:          "northpole",
		Name:              "unittestVM",
		ResourceGroupName: "unittestResourceGroup",
		SubscriptionID:    "96854241-60c7-426d-9a27-3fdeec8957f4",
	}

	tests := []struct {
		name   string
		config Config
		verify func([]multistep.Step, *testing.T)
	}{
		{
			name:   "Source FromScrath creates empty disk",
			config: Config{FromScratch: true},
			verify: func(steps []multistep.Step, _ *testing.T) {
				for _, s := range steps {
					if s, ok := s.(*StepCreateNewDisk); ok {
						if s.SourceDiskResourceID == "" &&
							s.PlatformImage == nil {
							return
						}
						t.Errorf("found misconfigured StepCreateNewDisk: %+v", s)
					}
				}
				t.Error("did not find a StepCreateNewDisk")
			}},
		{
			name:   "Source Platform image disk creation",
			config: Config{Source: "publisher:offer:sku:version", sourceType: sourcePlatformImage},
			verify: func(steps []multistep.Step, _ *testing.T) {
				for _, s := range steps {
					if s, ok := s.(*StepCreateNewDisk); ok {
						if s.SourceDiskResourceID == "" &&
							s.PlatformImage != nil &&
							s.PlatformImage.Publisher == "publisher" {
							return
						}
						t.Errorf("found misconfigured StepCreateNewDisk: %+v", s)
					}
				}
				t.Error("did not find a StepCreateNewDisk")
			}},
		{
			name:   "Source Platform image with version latest adds StepResolvePlatformImageVersion",
			config: Config{Source: "publisher:offer:sku:latest", sourceType: sourcePlatformImage},
			verify: func(steps []multistep.Step, _ *testing.T) {
				for _, s := range steps {
					if s, ok := s.(*StepResolvePlatformImageVersion); ok {
						if s.PlatformImage != nil &&
							s.Location == info.Location {
							return
						}
						t.Errorf("found misconfigured StepResolvePlatformImageVersion: %+v", s)
					}
				}
				t.Error("did not find a StepResolvePlatformImageVersion")
			}},
		{
			name:   "Source Disk adds correct disk creation",
			config: Config{Source: "diskresourceid", sourceType: sourceDisk},
			verify: func(steps []multistep.Step, _ *testing.T) {
				for _, s := range steps {
					if s, ok := s.(*StepCreateNewDisk); ok {
						if s.SourceDiskResourceID == "diskresourceid" &&
							s.PlatformImage == nil {
							return
						}
						t.Errorf("found misconfigured StepCreateNewDisk: %+v", s)
					}
				}
				t.Error("did not find a StepCreateNewDisk")
			}},
		{
			name:   "Source disk adds StepVerifySourceDisk",
			config: Config{Source: "diskresourceid", sourceType: sourceDisk},
			verify: func(steps []multistep.Step, _ *testing.T) {
				for _, s := range steps {
					if s, ok := s.(*StepVerifySourceDisk); ok {
						if s.SourceDiskResourceID == "diskresourceid" &&
							s.Location == info.Location &&
							s.SubscriptionID == info.SubscriptionID {
							return
						}
						t.Errorf("found misconfigured StepVerifySourceDisk: %+v", s)
					}
				}
				t.Error("did not find a StepVerifySourceDisk")
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMetadataStub(func() { // ensure that values are taken from info, instead of retrieved again
				got := buildsteps(tt.config, info)
				tt.verify(got, t)
			})
		})
	}
}
