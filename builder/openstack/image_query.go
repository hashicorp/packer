package openstack

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/packer/packer"
)

const (
	descendingSort = "desc"
	createdAtKey   = "created_at"
)

// Retrieve the specific ImageDateFilter using the exported const from images
func getDateFilter(s string) (images.ImageDateFilter, error) {
	filters := []images.ImageDateFilter{
		images.FilterGT,
		images.FilterGTE,
		images.FilterLT,
		images.FilterLTE,
		images.FilterNEQ,
		images.FilterEQ,
	}

	for _, filter := range filters {
		if string(filter) == s {
			return filter, nil
		}
	}

	var badFilter images.ImageDateFilter
	return badFilter, fmt.Errorf("No valid ImageDateFilter found for %s", s)
}

// Retrieve the specific ImageVisibility using the exported const from images
func getImageVisibility(s string) (images.ImageVisibility, error) {
	visibilities := [...]images.ImageVisibility{
		images.ImageVisibilityPublic,
		images.ImageVisibilityPrivate,
		images.ImageVisibilityCommunity,
		images.ImageVisibilityShared,
	}

	for _, visibility := range visibilities {
		if string(visibility) == s {
			return visibility, nil
		}
	}

	var nilVisibility images.ImageVisibility
	return nilVisibility, fmt.Errorf("No valid ImageVisilibility found for %s", s)
}

// Retrieve the specific ImageVisibility using the exported const from images
func getImageStatus(s string) (images.ImageStatus, error) {
	activeStatus := images.ImageStatusActive
	if string(activeStatus) == s {
		return activeStatus, nil
	}

	var nilStatus images.ImageStatus
	return nilStatus, fmt.Errorf("No valid ImageVisilibility found for %s", s)
}

// Allows construction of all fields from ListOpts using the "q" tags and
// type detection to set all fields within a provided ListOpts struct
func buildImageFilters(input map[string]interface{}, listOpts *images.ListOpts) *packer.MultiError {

	// fill each field in the ListOpts based on tag/type
	metaOpts := reflect.Indirect(reflect.ValueOf(listOpts))

	multiErr := packer.MultiError{}

	for i := 0; i < metaOpts.Type().NumField(); i++ {
		vField := metaOpts.Field(i)
		tField := metaOpts.Type().Field(i)
		fieldName := tField.Name
		key := metaOpts.Type().Field(i).Tag.Get("q")

		// get key from the map and set values if they exist
		if val, exists := input[key]; exists && vField.CanSet() {
			switch vField.Kind() {

			// Handles integer types used in ListOpts
			case reflect.Int64, reflect.Int:
				iVal, err := strconv.Atoi(val.(string))
				if err != nil {
					multierror.Append(err, multiErr.Errors...)
					continue
				}

				if vField.Kind() == reflect.Int {
					vField.Set(reflect.ValueOf(iVal))
				} else {
					var i64Val int64
					i64Val = int64(iVal)
					vField.Set(reflect.ValueOf(i64Val))
				}

			// Handles string and types using string
			case reflect.String:
				switch vField.Type() {
				default:
					vField.Set(reflect.ValueOf(val))

				case reflect.TypeOf(images.ImageVisibility("")):
					iv, err := getImageVisibility(val.(string))
					if err != nil {
						multierror.Append(err, multiErr.Errors...)
						continue
					}
					vField.Set(reflect.ValueOf(iv))

				case reflect.TypeOf(images.ImageStatus("")):
					is, err := getImageStatus(val.(string))
					if err != nil {
						multierror.Append(err, multiErr.Errors...)
						continue
					}
					vField.Set(reflect.ValueOf(is))
				}

			default:
				multierror.Append(
					fmt.Errorf("Unsupported kind %s", vField.Kind()),
					multiErr.Errors...)
			}

		} else if fieldName == reflect.TypeOf(listOpts.CreatedAtQuery).Name() ||
			fieldName == reflect.TypeOf(listOpts.UpdatedAtQuery).Name() {
			// Handles ImageDateQuery types

			query, err := dateToImageDateQuery(key, val.(string))
			if err != nil {
				multierror.Append(err, multiErr.Errors...)
				continue
			}

			vField.Set(reflect.ValueOf(query))

		} else if fieldName == reflect.TypeOf(listOpts.Tags).Name() {
			// Handles "tags" case and processes as slice of string

			if val, exists := input["tags"]; exists && vField.CanSet() {
				vField.Set(reflect.ValueOf(val))
			}
		}
	}

	return &multiErr
}

// Apply most recent filtering logic to ListOpts where user has filled fields.
// This does not check whether both are filled. Allow OpenStack to determine which to use.
// It is suggested that users use the newest sort field
// See https://developer.openstack.org/api-ref/image/v2/
func applyMostRecent(listOpts *images.ListOpts) {
	// Apply to old sorting properties if user used them. This overwrites previous values.
	// The docs don't seem to mention more than one field being allowed here and how they would be
	listOpts.SortDir = descendingSort
	listOpts.SortKey = createdAtKey

	// Apply to new sorting property.
	if listOpts.Sort != "" {
		listOpts.Sort = fmt.Sprintf("%s:%s,%s", createdAtKey, descendingSort, listOpts.Sort)
	} else {
		listOpts.Sort = fmt.Sprintf("%s:%s", createdAtKey, descendingSort)
	}

	return
}

// Converts a given date entry to ImageDateQuery for use in ListOpts
func dateToImageDateQuery(val string, key string) (*images.ImageDateQuery, error) {
	q := new(images.ImageDateQuery)
	sep := ":"
	entries := strings.Split(val, sep)

	if len(entries) > 3 {
		filter, err := getDateFilter(entries[0])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse date filter for %s", key)
		} else {
			q.Filter = filter
		}

		dateSubstr := val[len(entries[0])+1:]
		date, err := time.Parse(time.RFC3339, dateSubstr)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse date format for %s.\nDate: %s.\nError: %s",
				key,
				dateSubstr,
				err.Error())
		} else {
			q.Date = date
		}

		return q, nil
	}

	return nil, fmt.Errorf("Incorrect date query format for %s", key)
}
