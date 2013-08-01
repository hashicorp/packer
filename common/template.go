package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"reflect"
	"text/template"
)

type traverseFunc func(string, string) (string, bool)

// CheckTemplates verifies that all the string values in the given
// configuration struct are valid templates. If not, an error is
// returned.
func CheckTemplates(i interface{}) *packer.MultiError {
	v := reflect.ValueOf(i).Elem()
	if !v.CanAddr() {
		panic("Arg to CheckTemplates isn't addressable")
	}

	if v.Kind() != reflect.Struct {
		panic("Arg to CheckTemplates must be a struct")
	}

	errs := make([]error, 0)

	f := func(n string, s string) (string, bool) {
		_, err := template.New("field").Parse(s)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("%s is not a valid template: %s", n, err))
		}

		return "", false
	}

	traverseStructStrings("", v, f)
	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func ProcessTemplates(i interface{}) *packer.MultiError {
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
