package openstack

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/mitchellh/mapstructure"
)

func TestGetImageFilter(t *testing.T) {
	passedExpectedMap := map[string]images.ImageDateFilter{
		"gt":  images.FilterGT,
		"gte": images.FilterGTE,
		"lt":  images.FilterLT,
		"lte": images.FilterLTE,
		"neq": images.FilterNEQ,
		"eq":  images.FilterEQ,
	}

	for passed, expected := range passedExpectedMap {
		filter, err := getDateFilter(passed)
		if err != nil {
			t.Errorf("Passed %s, received error: %s", passed, err.Error())
		} else if filter != expected {
			t.Errorf("Expected %s, got %s", expected, filter)
		}
	}
}

func TestBuildImageFilter(t *testing.T) {
	testOpts := images.ListOpts{}

	filters := map[string]string{
		"limit":      "3",
		"name":       "Ubuntu 16.04",
		"visibility": "public",
		"status":     "active",
		"size_min":   "0",
		"sort":       "created_at:desc",
	}

	multiErr := buildImageFilters(filters, &testOpts)

	if multiErr != nil {
		for _, err := range multiErr.Errors {
			t.Error(err)
		}
	}

	if testOpts.Limit != 3 {
		t.Errorf("Limit did not parse correctly: %d", testOpts.Limit)
	}

	if testOpts.Name != filters["name"] {
		t.Errorf("Name did not parse correctly: %s", filters["name"])
	}

	var visibility images.ImageVisibility = "public"
	if testOpts.Visibility != visibility {
		t.Errorf("Visibility did not parse correctly")
	}

	var imageStatus images.ImageStatus = "active"
	if testOpts.Status != imageStatus {
		t.Errorf("Image status did not parse correctly: %s", testOpts.Status)
	}

	if testOpts.SizeMin != 0 {
		t.Errorf("Size min did not parse correctly: %s", filters["size_min"])
	}

	if testOpts.Sort != filters["sort"] {
		t.Errorf("Sort did not parse correctly: %s", filters["sort"])
	}
}

func TestApplyMostRecent(t *testing.T) {
	testSortEmptyOpts := images.ListOpts{
		Name:    "RHEL 7.0",
		SizeMin: 0,
	}

	testSortFilledOpts := images.ListOpts{
		Name:    "Ubuntu 16.04",
		SizeMin: 0,
		Sort:    "tags:ubuntu",
	}

	applyMostRecent(&testSortEmptyOpts)

	if testSortEmptyOpts.Sort != "created_at:desc" {
		t.Errorf("Error applying most recent filter: sort")
	}

	if testSortEmptyOpts.SortDir != "desc" || testSortEmptyOpts.SortKey != "created_at" {
		t.Errorf("Error applying most recent filter: sort_dir/sort_key:\n{sort_dir: %s, sort_key: %s}",
			testSortEmptyOpts.SortDir, testSortEmptyOpts.SortKey)
	}

	applyMostRecent(&testSortFilledOpts)

	if testSortFilledOpts.Sort != "created_at:desc,tags:ubuntu" {
		t.Errorf("Error applying most recent filter: sort")
	}

	if testSortFilledOpts.SortDir != "desc" || testSortFilledOpts.SortKey != "created_at" {
		t.Errorf("Error applying most recent filter: sort_dir/sort_key:\n{sort_dir: %s, sort_key: %s}",
			testSortFilledOpts.SortDir, testSortFilledOpts.SortKey)
	}
}

func TestDateToImageDateQuery(t *testing.T) {
	tests := [][2]string{
		{"gt:2012-11-01T22:08:41+00:00", "created_at"},
	}

	for _, test := range tests {
		if _, err := dateToImageDateQuery(test[0], test[1]); err != nil {
			t.Error(err)
		}
	}
}

func TestImageFilterOptionsDecode(t *testing.T) {
	opts := ImageFilterOptions{}
	input := map[string]interface{}{
		"most_recent": true,
		"filters": map[string]interface{}{
			"visibility": "protected",
			"tag":        "prod",
			"name":       "ubuntu 16.04",
		},
	}
	err := mapstructure.Decode(input, &opts)
	if err != nil {
		t.Error("Did not successfully generate ImageFilterOptions from %v. Contains %v", input, opts)
	}
}
