package chroot

import (
	"reflect"
	"strings"
	"testing"
)

func TestSharedImageGalleryDestination_ResourceID(t *testing.T) {
	sigd := SharedImageGalleryDestination{
		ResourceGroup: "ResourceGroup",
		GalleryName:   "GalleryName",
		ImageName:     "ImageName",
		ImageVersion:  "ImageVersion",
	}
	want := "/subscriptions/SubscriptionID/resourceGroups/ResourceGroup/providers/Microsoft.Compute/galleries/GalleryName/images/ImageName/versions/ImageVersion"
	if got := sigd.ResourceID("SubscriptionID"); !strings.EqualFold(got, want) {
		t.Errorf("SharedImageGalleryDestination.ResourceID() = %v, want %v", got, want)
	}
}

func TestSharedImageGalleryDestination_Validate(t *testing.T) {
	type fields struct {
		ResourceGroup     string
		GalleryName       string
		ImageName         string
		ImageVersion      string
		TargetRegions     []TargetRegion
		ExcludeFromLatest bool
	}
	tests := []struct {
		name      string
		fields    fields
		wantErrs  []string
		wantWarns []string
	}{
		{
			name: "complete",
			fields: fields{
				ResourceGroup: "ResourceGroup",
				GalleryName:   "GalleryName",
				ImageName:     "ImageName",
				ImageVersion:  "0.1.2",
				TargetRegions: []TargetRegion{
					TargetRegion{
						Name:               "region1",
						ReplicaCount:       5,
						StorageAccountType: "Standard_ZRS",
					},
					TargetRegion{
						Name:               "region2",
						ReplicaCount:       3,
						StorageAccountType: "Standard_LRS",
					},
				},
				ExcludeFromLatest: true,
			},
		},
		{
			name: "warn if target regions not specified",
			fields: fields{
				ResourceGroup: "ResourceGroup",
				GalleryName:   "GalleryName",
				ImageName:     "ImageName",
				ImageVersion:  "0.1.2",
			},
			wantWarns: []string{"sigdest.target_regions is empty; image will only be available in the region of the gallery"},
		},
		{
			name: "version format",
			wantErrs: []string{
				"sigdest.image_version should match '^[0-9]+\\.[0-9]+\\.[0-9]+$'",
			},
			fields: fields{
				ResourceGroup: "ResourceGroup",
				GalleryName:   "GalleryName",
				ImageName:     "ImageName",
				ImageVersion:  "0.1.2alpha",
				TargetRegions: []TargetRegion{
					TargetRegion{
						Name:               "region1",
						ReplicaCount:       5,
						StorageAccountType: "Standard_ZRS",
					},
					TargetRegion{
						Name:               "region2",
						ReplicaCount:       3,
						StorageAccountType: "Standard_LRS",
					},
				},
				ExcludeFromLatest: true,
			},
		},
		{
			name: "required fields",
			wantErrs: []string{
				"sigdest.resource_group is required",
				"sigdest.gallery_name is required",
				"sigdest.image_name is required",
				"sigdest.image_version should match '^[0-9]+\\.[0-9]+\\.[0-9]+$'",
			},
			wantWarns: []string{"sigdest.target_regions is empty; image will only be available in the region of the gallery"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigd := &SharedImageGalleryDestination{
				ResourceGroup:     tt.fields.ResourceGroup,
				GalleryName:       tt.fields.GalleryName,
				ImageName:         tt.fields.ImageName,
				ImageVersion:      tt.fields.ImageVersion,
				TargetRegions:     tt.fields.TargetRegions,
				ExcludeFromLatest: tt.fields.ExcludeFromLatest,
			}
			gotErrs, gotWarns := sigd.Validate("sigdest")

			var gotStrErrs []string
			if gotErrs != nil {
				gotStrErrs = make([]string, len(gotErrs))
				for i, e := range gotErrs {
					gotStrErrs[i] = e.Error()
				}
			}

			if !reflect.DeepEqual(gotStrErrs, tt.wantErrs) {
				t.Errorf("SharedImageGalleryDestination.Validate() gotErrs = %q, want %q", gotStrErrs, tt.wantErrs)
			}
			if !reflect.DeepEqual(gotWarns, tt.wantWarns) {
				t.Errorf("SharedImageGalleryDestination.Validate() gotWarns = %q, want %q", gotWarns, tt.wantWarns)
			}
		})
	}
}
