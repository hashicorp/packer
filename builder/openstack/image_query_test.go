package openstack

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/mitchellh/mapstructure"
)

func TestImageFilterOptionsDecode(t *testing.T) {
	opts := ImageFilterOptions{}
	input := map[string]interface{}{
		"most_recent": true,
		"filters": map[string]interface{}{
			"visibility": "protected",
			"tag":        []string{"prod", "ready"},
			"name":       "ubuntu 16.04",
			"owner":      "tcarrio",
		},
	}
	err := mapstructure.Decode(input, &opts)
	if err != nil {
		t.Errorf("Did not successfully generate ImageFilterOptions from %v.\nContains %v", input, opts)
	}
}

// This test case confirms that only allowed fields will be set to values
// The checked values are non-nil for their target type
func TestBuildImageFilter(t *testing.T) {
	testOpts := images.ListOpts{}

	filters := map[string]interface{}{
		"limit":      "3",
		"name":       "Ubuntu 16.04",
		"visibility": "public",
		"status":     "active",
		"size_min":   "25",
		"sort":       "created_at:desc",
		"tags":       []string{"prod", "ready"},
	}

	// copy of original filters to pass to build function
	passedFilters := make(map[string]interface{})
	for k, v := range filters {
		passedFilters[k] = v
	}

	buildImageFilters(passedFilters, &testOpts)

	if testOpts.Limit != 0 {
		t.Errorf("Limit was parsed: %d", testOpts.Limit)
	}

	if testOpts.Name != filters["name"] {
		t.Errorf("Name did not parse correctly: %s", testOpts.Name)
	}

	if testOpts.Visibility != images.ImageVisibilityPublic {
		t.Errorf("Visibility did not parse correctly: %v", testOpts.Visibility)
	}

	if testOpts.Status != images.ImageStatusActive {
		t.Errorf("Image status did not parse correctly: %s", testOpts.Status)
	}

	if testOpts.SizeMin != 0 {
		t.Errorf("Size min was parsed: %d", testOpts.SizeMin)
	}

	if len(testOpts.Sort) > 0 {
		t.Errorf("Sort was parsed: %s", testOpts.Sort)
	}
}

// This test case confirms that invalid filter input are caught and do not result in a panic
func TestInvalidFilterInput(t *testing.T) {

	testOpts := images.ListOpts{}

	filters := map[string]interface{}{
		"tags":          "prod", // supposed to be a []string
		"owner":         12345,  // supposed to be a string
		"invalid_field": 0,      // not a valid field in ListOpts
	}

	numFields := len(filters)

	multiErr := buildImageFilters(filters, &testOpts)
	if len(multiErr.Errors) != numFields {
		t.Errorf("Failed to catch all %d invalid types/fields in filters", numFields)
		for _, err := range multiErr.Errors {
			t.Log(err.Error())
		}
	}
}

func TestApplyMostRecent(t *testing.T) {
	testSortOpts := images.ListOpts{
		Name: "RHEL 7.0",
		Tags: []string{"prod", "ready"},
	}

	applyMostRecent(&testSortOpts)

	if testSortOpts.Sort != "created_at:desc" {
		t.Errorf("Error applying most recent filter: sort")
	}
}
