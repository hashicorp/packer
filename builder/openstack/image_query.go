package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/packer"
	"reflect"
)

const (
	mostRecentSortDir = "desc"
	mostRecentSortKey = "created_at"
)

var validFields = map[string]string{
	"Name":       "name",
	"Visibility": "visibility",
	"Owner":      "owner",
	"Tags":       "tags",
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
	return nilVisibility, fmt.Errorf("No valid ImageVisibility found for %s", s)
}

// Allows construction of all supported fields from ListOpts
func buildImageFilters(input map[string]interface{}, listOpts *images.ListOpts) *packer.MultiError {

	// fill each field in the ListOpts based on tag/type
	metaOpts := reflect.Indirect(reflect.ValueOf(listOpts))
	multiErr := packer.MultiError{}

	for i := 0; i < metaOpts.Type().NumField(); i++ {
		vField := metaOpts.Field(i)
		fieldName := metaOpts.Type().Field(i).Name

		// check the valid fields map and whether we can set this field
		if key, exists := validFields[fieldName]; exists {
			if !vField.CanSet() {
				multiErr.Errors = append(multiErr.Errors, fmt.Errorf("Unsettable field: %s", fieldName))
				continue
			}

			// check that this key was provided by the user, then set the field and have compatible types
			if val, exists := input[key]; exists {

				switch key {
				case "owner", "name", "tags":

					if valType := reflect.TypeOf(val); valType != vField.Type() {
						multiErr.Errors = append(multiErr.Errors,
							fmt.Errorf("Invalid type '%v' for field %s",
								valType,
								fieldName,
							))
						continue
					}
					vField.Set(reflect.ValueOf(val))

				case "visibility":
					visibility, err := getImageVisibility(val.(string))
					if err != nil {
						multiErr.Errors = append(multiErr.Errors, err)
						continue
					}
					vField.Set(reflect.ValueOf(visibility))

				default:
					multiErr.Errors = append(multiErr.Errors,
						fmt.Errorf("Unsupported filter key provided: %s", key))
				}
			}
		}
	}

	// Set defaults for status and member_status
	listOpts.Status = images.ImageStatusActive
	listOpts.MemberStatus = images.ImageMemberStatusAccepted

	return &multiErr
}

// Apply most recent filtering logic to ListOpts where user has filled fields.
// See https://developer.openstack.org/api-ref/image/v2/
func applyMostRecent(listOpts *images.ListOpts) {
	// Sort isn't supported through our API so there should be no existing values.
	// Overwriting .Sort is okay.
	listOpts.SortKey = mostRecentSortKey
	listOpts.SortDir = mostRecentSortDir
	listOpts.Limit = 1

	return
}
