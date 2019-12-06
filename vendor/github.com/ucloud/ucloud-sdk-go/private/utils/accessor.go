package utils

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// ValueAtPath will get struct attribute value by recursive
func ValueAtPath(v interface{}, path string) (interface{}, error) {
	components := strings.Split(path, ".")

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, errors.Errorf("object %#v is nil", v)
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		i, err := strconv.Atoi(components[0])
		if err != nil {
			return nil, errors.Errorf("path %s is invalid at index of array", path)
		}

		length := rv.Len()
		if i >= length {
			return nil, errors.Errorf("path %s is invalid, array has length %v, but got %v", path, length, i)
		}

		itemV := rv.Index(i)
		if !itemV.IsValid() {
			return nil, errors.Errorf("path %s is invalid for map", path)
		}

		if len(components) > 1 {
			return ValueAtPath(itemV.Interface(), strings.Join(components[1:], "."))
		}

		return itemV.Interface(), nil
	}

	if rv.Kind() == reflect.Map && !rv.IsNil() {
		itemV := rv.MapIndex(reflect.ValueOf(components[0]))
		if !itemV.IsValid() {
			return nil, errors.Errorf("path %s is invalid for map", path)
		}

		if len(components) > 1 {
			return ValueAtPath(itemV.Interface(), strings.Join(components[1:], "."))
		}

		return itemV.Interface(), nil
	}

	if rv.Kind() == reflect.Struct {
		itemV := rv.FieldByNameFunc(func(s string) bool {
			return strings.ToLower(s) == strings.ToLower(components[0])
		})

		if !itemV.IsValid() {
			return nil, errors.Errorf("path %s is invalid for struct", path)
		}

		if len(components) > 1 {
			return ValueAtPath(itemV.Interface(), strings.Join(components[1:], "."))
		}

		return itemV.Interface(), nil
	}

	return nil, errors.Errorf("object %#v is invalid, need map or struct", v)
}
