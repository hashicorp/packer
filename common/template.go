package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"reflect"
	"text/template"
)

type traverseFunc func(string, string) (string, bool)

// ConfigTemplate processes your entire configuration struct and processes
// all strings through the Golang text/template processor. This exposes
// common functions to all strings within Packer without any extra effort
// by the implementor.
type ConfigTemplate struct {
	v reflect.Value
}

// NewConfigTemplate will return a new configuration template processor
// for the given interface. The interface passed in should generally be
// a pointer to your configuration struct, because ConfigTemplate will
// modify data in-place.
func NewConfigTemplate(i interface{}) (*ConfigTemplate, error) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Interface && v.Kind() != reflect.Ptr {
		return nil, errors.New("Interface should be an interface or pointer.")
	}

	v = v.Elem()
	if !v.CanAddr() {
		return nil, errors.New("Interface isn't addressable")
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("Interface must be a struct")
	}

	return &ConfigTemplate{
		v: v,
	}, nil
}

// Check verifies that all the string values in the given
// configuration struct are valid templates. If not, an error is
// returned.
func (ct *ConfigTemplate) Check() error {
	errs := make([]error, 0)

	f := func(n string, s string) (string, bool) {
		_, err := template.New("field").Parse(s)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("%s is not a valid template: %s", n, err))
		}

		return "", false
	}

	traverseStructStrings("", ct.v, f)
	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func traverseMapStrings(n string, v reflect.Value, f traverseFunc) {
	n = n + "."
	for _, k := range v.MapKeys() {
		if k.Kind() != reflect.String {
			return
		}

		kv := v.MapIndex(k)
		fieldName := n + k.Interface().(string)
		traverseValue(fieldName, k, f)
		traverseValue(fieldName, kv, f)
	}
}

func traverseSliceStrings(n string, v reflect.Value, f traverseFunc) {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		fieldName := fmt.Sprintf("%s[%d]", n, i)
		traverseValue(fieldName, elem, f)
	}
}

func traverseStructStrings(n string, v reflect.Value, f traverseFunc) {
	n = n + "."
	vt := v.Type()
	for i := 0; i < vt.NumField(); i++ {
		field := v.FieldByIndex([]int{i})
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		sf := vt.Field(i)
		fieldName := n + sf.Name
		traverseValue(fieldName, field, f)
	}
}

func traverseValue(n string, v reflect.Value, f traverseFunc) (string, bool) {
	switch v.Kind() {
	case reflect.Map:
		traverseMapStrings(n, v, f)
	case reflect.Struct:
		traverseStructStrings(n, v, f)
	case reflect.Slice:
		traverseSliceStrings(n, v, f)
	case reflect.String:
		return f(n, v.Interface().(string))
	}

	return "", false
}
