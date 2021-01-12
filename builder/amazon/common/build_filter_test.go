package common

import (
	"testing"
)

func TestStepSourceAmiInfo_BuildFilter(t *testing.T) {
	filter_key := "name"
	filter_value := "foo"
	filter_key2 := "name2"
	filter_value2 := "foo2"

	inputFilter := map[string]string{filter_key: filter_value, filter_key2: filter_value2}
	outputFilter := buildEc2Filters(inputFilter)

	// deconstruct filter back into things we can test
	foundMap := map[string]bool{filter_key: false, filter_key2: false}
	for _, filter := range outputFilter {
		for key, value := range inputFilter {
			if *filter.Name == key && *filter.Values[0] == value {
				foundMap[key] = true
			}
		}
	}

	for k, v := range foundMap {
		if !v {
			t.Fatalf("Fail: should have found value for key: %s", k)
		}
	}
}
