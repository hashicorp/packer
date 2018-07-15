package openstack

import (
	"testing"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

func TestGetImageFilter(t *testing.T) {
	passedExpectedMap := map[string]images.ImageDateFilter{
		"gt":   images.FilterGT,
		"gte":  images.FilterGTE,
		"lt":   images.FilterLT,
		"lte":  images.FilterLTE,
		"neq":  images.FilterNEQ,
		"eq":   images.FilterEQ,
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
		"limit": "3",
		"name": "Ubuntu 16.04",
		"visibility": "public",
		"image_status": "active",
		"size_min": "0",
		"sort": "created_at:desc",
	}

	multiErr := buildImageFilters(filters, &testOpts)

	if multiErr != nil {
		for _, err := range multiErr.Errors {
			t.Error(err)
		}
	}

	if testOpts.Limit != 3 {
		t.Errorf("Limit did not parse correctly")
	}

	if testOpts.Name != filters["name"] {
		t.Errorf("Name did not parse correctly")
	}

	var visibility images.ImageVisibility = "public"
	if testOpts.Visibility != visibility {
		t.Errorf("Visibility did not parse correctly")
	}

	var imageStatus images.ImageStatus = "active"
	if testOpts.Status != imageStatus {
		t.Errorf("Image status did not parse correctly")
	}

	if testOpts.SizeMin != 0 {
		t.Errorf("Size min did not parse correctly")
	}

	if testOpts.Sort != filters["sort"] {
		t.Errorf("Limit did not parse correctly")
	}
}

func TestApplyMostRecent(t *testing.T) {
	testOpts := images.ListOpts{
		Name: "RHEL 7.0",
		SizeMin: 0,
	}

	applyMostRecent(&testOpts)

	if testOpts.Sort != "created_at:desc" {
		t.Errorf("Error applying most recent filter: sort")
	}

	if testOpts.SortDir != "desc" || testOpts.SortKey != "created_at" {
		t.Errorf("Error applying most recent filter: sort_dir/sort_key")
	}
}

func TestDateToImageDateQuery(t *testing.T) {
	tests := [][2]string{
		{"2006-01-02T15:04:05Z07:00", "created_at"},
	}

	for _, test := range tests {
		if _, err := dateToImageDateQuery(&test[0], &test[1]); err != nil {
			t.Error(err)
		}
	}
}