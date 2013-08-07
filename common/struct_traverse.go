package common

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type TraverseFunc func(string, string) string

func TraverseStructStrings(i interface{}, f TraverseFunc) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Interface && v.Kind() != reflect.Ptr {
		return errors.New("Interface should be an interface or pointer.")
	}

	v = v.Elem()
	if !v.CanAddr() {
		return errors.New("Interface isn't addressable")
	}

	if v.Kind() != reflect.Struct {
		return errors.New("Interface must be a struct")
	}

	traverseStructStrings("", v, f)
	return nil
}

func traverseMapStrings(n string, v reflect.Value, f TraverseFunc) {
	n = n + "."
	for _, k := range v.MapKeys() {
		if k.Kind() != reflect.String {
			return
		}

		kv := v.MapIndex(k)
		fieldName := n + k.Interface().(string)
		newK, kRep := traverseValue(fieldName+" (key)", k, f)
		newV, vRep := traverseValue(fieldName, kv, f)

		var replaceKey, replaceValue reflect.Value
		if vRep {
			replaceValue = reflect.ValueOf(newV)
		} else {
			replaceValue = kv
		}

		if kRep {
			v.SetMapIndex(k, reflect.Zero(kv.Type()))
			replaceKey = reflect.ValueOf(newK)
		} else {
			replaceKey = k
		}

		v.SetMapIndex(replaceKey, replaceValue)
	}
}

func traverseSliceStrings(n string, v reflect.Value, f TraverseFunc) {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		fieldName := fmt.Sprintf("%s[%d]", n, i)
		if r, do := traverseValue(fieldName, elem, f); do {
			elem.SetString(r)
		}
	}
}

func traverseStructStrings(n string, v reflect.Value, f TraverseFunc) {
	if n != "" {
		n = n + "."
	}

	vt := v.Type()
	for i := 0; i < vt.NumField(); i++ {
		field := v.FieldByIndex([]int{i})
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		// If the field isn't exported, then ignore it.
		if !field.CanSet() {
			continue
		}

		// Determine the field name. By default it is just the lowercase
		// field name, but if a mapstructure field name is specified,
		// prefer that.
		sf := vt.Field(i)
		fieldName := strings.ToLower(sf.Name)
		mapstructureTag := sf.Tag.Get("mapstructure")
		if mapstructureTag != "" {
			commaIdx := strings.Index(mapstructureTag, ",")
			if commaIdx == -1 {
				commaIdx = len(mapstructureTag)
			}

			fieldName = mapstructureTag[0:commaIdx]
		}

		fieldName = n + fieldName
		if r, do := traverseValue(fieldName, field, f); do {
			field.SetString(r)
		}
	}
}

func traverseValue(n string, v reflect.Value, f TraverseFunc) (string, bool) {
	switch v.Kind() {
	case reflect.Map:
		traverseMapStrings(n, v, f)
	case reflect.Struct:
		traverseStructStrings(n, v, f)
	case reflect.Slice:
		traverseSliceStrings(n, v, f)
	case reflect.String:
		return f(n, v.Interface().(string)), true
	}

	return "", false
}
