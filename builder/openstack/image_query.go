package openstack

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"strconv"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/go-multierror"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

const (
	descendingSort = "desc"
	createdAtKey   = "created_at"
)

// Retrieve the specific ImageDateFilter using the exported const from images
func getDateFilter(s string) (images.ImageDateFilter, error) {
	filters := [...]images.ImageDateFilter{
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

	return images.ImageDateFilter(nil), fmt.Errorf("No ImageDateFilter found for %s", s)
}

// Allows construction of all fields from ListOpts using the "q" tags and
// type detection to set all fields within a provided ListOpts struct
func buildImageFilters(input map[string]string, listOpts *images.ListOpts) *packer.MultiError {

	// fill each field in the ListOpts based on tag/type
	metaOpts := reflect.ValueOf(listOpts).Elem()

	multiErr := packer.MultiError{}

	for i := 0; i < metaOpts.Type().NumField(); i++ {
		vField := metaOpts.Field(i)
		tField := metaOpts.Type().Field(i)
		fieldName := tField.Name
		key := metaOpts.Type().Field(i).Tag.Get("q")

		// get key from the map and set values if they exist
		if val, exists := input[key]; exists && vField.CanSet() {
			switch vField.Kind() {

			case reflect.Int64:
				iVal, err := strconv.Atoi(val)
				if err != nil {
					multierror.Append(err, multiErr.Errors...)
				} else {
					vField.Set(reflect.ValueOf(iVal))
				}

			case reflect.String:
				vField.Set(reflect.ValueOf(val))

			case reflect.Slice:
				typeOfSlice := reflect.TypeOf(vField).Elem()
				fieldArray := reflect.MakeSlice(reflect.SliceOf(typeOfSlice), 0, 0)
				for _, s := range strings.Split(val, ",") {
					if len(s) > 0 {
						fieldArray = reflect.Append(fieldArray, reflect.ValueOf(s))
					}
				}
				vField.Set(fieldArray)

			default:
				multierror.Append(
					fmt.Errorf("Unsupported struct type %s", vField.Type().Name),
					multiErr.Errors...)
			}

		} else if fieldName == reflect.TypeOf(images.ListOpts{}.CreatedAtQuery).Name() ||
			fieldName == reflect.TypeOf(images.ListOpts{}.UpdatedAtQuery).Name() {
			// get ImageDateQuery from string and set to this field
			query, err := dateToImageDateQuery(&key, &val)
			if err != nil {
				multierror.Append(err, multiErr.Errors...)
				continue
			}
			vField.Set(reflect.ValueOf(query))
		}
	}

	return &multiErr
}

// Apply most recent filtering logic to ListOpts where user has filled fields.
// This does not check whether both are filled. Allow OpenStack to determine which to use.
// It is suggested that users use the newest sort field
// See https://developer.openstack.org/api-ref/image/v2/
func applyMostRecent(listOpts *images.ListOpts) {
	// apply to old sorting properties if user used them. This overwrites previous values?
	if listOpts.SortDir == "" && listOpts.SortKey != "" {
		listOpts.SortDir = descendingSort
		listOpts.SortKey = createdAtKey
	}

	// apply to new sorting property
	if listOpts.Sort != "" {
		listOpts.Sort = fmt.Sprintf("%s:%s,%s", createdAtKey, descendingSort, listOpts.Sort)
	} else {
		listOpts.Sort = fmt.Sprintf("%s:%s", createdAtKey, descendingSort)
	}

	return
}

// Converts a given date entry to ImageDateQuery for use in ListOpts
func dateToImageDateQuery(val *string, key *string) (*images.ImageDateQuery, error) {
	q := images.ImageDateQuery{}
	sep := ":"
	entries := strings.Split(*val, sep)

	if len(entries) > 3 {
		filter, err := getDateFilter(entries[0])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse date filter for %s", key)
		} else {
			q.Filter = filter
		}

		date, err := time.Parse((*val)[len(entries[0]):], time.RFC3339)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse date format for %s", key)
		} else {
			q.Date = date
		}

		return &q, nil
	}

	return nil, fmt.Errorf("Incorrect date query format for %s", key)
}
