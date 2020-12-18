package openstack

import (
	"os"
	"regexp"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/mitchellh/mapstructure"
)

func init() {
	// Clear out the openstack env vars so they don't
	// affect our tests.
	os.Setenv("SDK_USERNAME", "")
	os.Setenv("SDK_PASSWORD", "")
	os.Setenv("SDK_PROVIDER", "")
}

func testRunConfig() *RunConfig {
	return &RunConfig{
		SourceImage: "abcd",
		Flavor:      "m1.small",

		Comm: communicator.Config{
			SSH: communicator.SSH{
				SSHUsername: "foo",
			},
		},
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := testRunConfig()
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_InstanceType(t *testing.T) {
	c := testRunConfig()
	c.Flavor = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SourceImage(t *testing.T) {
	c := testRunConfig()
	c.SourceImage = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testRunConfig()
	c.Comm.SSHPort = 0
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}

	c.Comm.SSHPort = 44
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}
}

func TestRunConfigPrepare_BlockStorage(t *testing.T) {
	c := testRunConfig()
	c.UseBlockStorageVolume = true
	c.VolumeType = "fast"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.VolumeType != "fast" {
		t.Fatalf("invalid value: %s", c.VolumeType)
	}

	c.AvailabilityZone = "RegionTwo"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.VolumeAvailabilityZone != "RegionTwo" {
		t.Fatalf("invalid value: %s", c.VolumeAvailabilityZone)
	}

	c.VolumeAvailabilityZone = "RegionOne"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.VolumeAvailabilityZone != "RegionOne" {
		t.Fatalf("invalid value: %s", c.VolumeAvailabilityZone)
	}

	c.VolumeName = "PackerVolume"
	if c.VolumeName != "PackerVolume" {
		t.Fatalf("invalid value: %s", c.VolumeName)
	}
}

func TestRunConfigPrepare_FloatingIPPoolCompat(t *testing.T) {
	c := testRunConfig()
	c.FloatingIPPool = "uuid1"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.FloatingIPNetwork != "uuid1" {
		t.Fatalf("invalid value: %s", c.FloatingIPNetwork)
	}

	c.FloatingIPNetwork = "uuid2"
	c.FloatingIPPool = "uuid3"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.FloatingIPNetwork != "uuid2" {
		t.Fatalf("invalid value: %s", c.FloatingIPNetwork)
	}
}

func TestRunConfigPrepare_ExternalSourceImageURL(t *testing.T) {
	c := testRunConfig()
	// test setting both ExternalSourceImageURL and SourceImage causes an error
	c.ExternalSourceImageURL = "http://example.com/image.qcow2"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	c = testRunConfig()
	// test setting both ExternalSourceImageURL and SourceImageName causes an error
	c.SourceImage = ""
	c.SourceImageName = "abcd"
	c.ExternalSourceImageURL = "http://example.com/image.qcow2"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	c = testRunConfig()
	// test neither setting SourceImage, SourceImageName or ExternalSourceImageURL causes an error
	c.SourceImage = ""
	c.SourceImageName = ""
	c.ExternalSourceImageURL = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}

	c = testRunConfig()
	// test setting only ExternalSourceImageURL passes
	c.SourceImage = ""
	c.SourceImageName = ""
	c.ExternalSourceImageURL = "http://example.com/image.qcow2"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	// test default values
	if c.ExternalSourceImageFormat != "qcow2" {
		t.Fatalf("ExternalSourceImageFormat should have been set to default: qcow2")
	}

	p := `packer_[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`
	if matches, _ := regexp.MatchString(p, c.SourceImageName); !matches {
		t.Fatalf("invalid format for SourceImageName: %s", c.SourceImageName)
	}

	c = testRunConfig()
	// test setting a filter passes
	c.SourceImage = ""
	c.SourceImageName = ""
	c.ExternalSourceImageURL = ""
	c.SourceImageFilters = ImageFilter{
		Filters: ImageFilterOptions{
			Name:       "Ubuntu 16.04",
			Visibility: "public",
			Owner:      "1234567890",
			Tags:       []string{"prod", "ready"},
			Properties: map[string]string{"os_distro": "ubuntu", "os_version": "16.04"},
		},
		MostRecent: true,
	}

	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("Should not error if everything but filter is empty: %s", err)
	}

}

// This test case confirms that only allowed fields will be set to values
// The checked values are non-nil for their target type
func TestBuildImageFilter(t *testing.T) {

	filters := ImageFilterOptions{
		Name:       "Ubuntu 16.04",
		Visibility: "public",
		Owner:      "1234567890",
		Tags:       []string{"prod", "ready"},
		Properties: map[string]string{"os_distro": "ubuntu", "os_version": "16.04"},
	}

	listOpts, err := filters.Build()
	if err != nil {
		t.Errorf("Building filter failed with: %s", err)
	}

	if listOpts.Name != "Ubuntu 16.04" {
		t.Errorf("Name did not build correctly: %s", listOpts.Name)
	}

	if listOpts.Visibility != images.ImageVisibilityPublic {
		t.Errorf("Visibility did not build correctly: %s", listOpts.Visibility)
	}

	if listOpts.Owner != "1234567890" {
		t.Errorf("Owner did not build correctly: %s", listOpts.Owner)
	}
}

func TestBuildBadImageFilter(t *testing.T) {
	filterMap := map[string]interface{}{
		"limit":    "3",
		"size_min": "25",
	}

	filters := ImageFilterOptions{}
	mapstructure.Decode(filterMap, &filters)
	listOpts, err := filters.Build()

	if err != nil {
		t.Errorf("Error returned processing image filter: %s", err.Error())
		return // we cannot trust listOpts to not cause unexpected behaviour
	}

	if listOpts.Limit == filterMap["limit"] {
		t.Errorf("Limit was parsed into ListOpts: %d", listOpts.Limit)
	}

	if listOpts.SizeMin != 0 {
		t.Errorf("SizeMin was parsed into ListOpts: %d", listOpts.SizeMin)
	}

	if listOpts.Sort != "created_at:desc" {
		t.Errorf("Sort was not applied: %s", listOpts.Sort)
	}

	if !filters.Empty() {
		t.Errorf("The filters should be empty due to lack of input")
	}
}

// Tests that the Empty method on ImageFilterOptions works as expected
func TestImageFiltersEmpty(t *testing.T) {
	filledFilters := ImageFilterOptions{
		Name:       "Ubuntu 16.04",
		Visibility: "public",
		Owner:      "1234567890",
		Tags:       []string{"prod", "ready"},
		Properties: map[string]string{"os_distro": "ubuntu", "os_version": "16.04"},
	}

	if filledFilters.Empty() {
		t.Errorf("Expected filled filters to be non-empty: %v", filledFilters)
	}

	emptyFilters := ImageFilterOptions{}

	if !emptyFilters.Empty() {
		t.Errorf("Expected default filter to be empty: %v", emptyFilters)
	}
}
