package common

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type traverseFunc func(string, string) string

// ConfigTemplate processes your entire configuration struct and processes
// all strings through the Golang text/template processor. This exposes
// common functions to all strings within Packer without any extra effort
// by the implementor.
type ConfigTemplate struct {
	BuilderVars map[string]string
	UserVars    map[string]string
	root        *template.Template
	t           map[string]*template.Template
	v           reflect.Value
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

	result := &ConfigTemplate{
		BuilderVars: make(map[string]string),
		UserVars:    make(map[string]string),
		t:           make(map[string]*template.Template),
		v:           v,
	}

	root := template.New("configTemplateRoot")
	root.Funcs(template.FuncMap{
		"builder":   result.Builder,
		"timestamp": templateTimestamp,
		"user":      result.User,
	})

	// Set the template root so we can have a place to store
	// our template data.
	result.root = root

	return result, nil
}

// Check verifies that all the string values in the given
// configuration struct are valid templates. If not, an error is
// returned.
func (ct *ConfigTemplate) Check() error {
	errs := make([]error, 0)

	f := func(n string, s string) string {
		t, err := ct.root.New(n).Parse(s)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("%s is not a valid template: %s", n, err))
		} else {
			ct.t[n] = t
		}

		return s
	}

	traverseStructStrings("", ct.v, f)
	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

// Process goes over all the string values in the structure and runs
// the template on each of them, modifying them in place.
func (ct *ConfigTemplate) Process() error {
	buf := new(bytes.Buffer)
	errs := make([]error, 0)

	f := func(n string, s string) string {
		t, ok := ct.t[n]
		if !ok {
			panic("template not found: " + n)
		}

		buf.Reset()
		err := t.Execute(buf, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing %s: %s", n, err))
		}

		return buf.String()
	}

	traverseStructStrings("", ct.v, f)
	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

// Builder is the function exposed as "builder" within the templates and
// looks up builder variables.
func (ct *ConfigTemplate) Builder(n string) (string, error) {
	result, ok := ct.BuilderVars[n]
	if !ok {
		return "", fmt.Errorf("uknown builder var: %s", n)
	}

	return result, nil
}

// User is the function exposed as "user" within the templates and
// looks up user variables.
func (ct *ConfigTemplate) User(n string) (string, error) {
	result, ok := ct.UserVars[n]
	if !ok {
		return "", fmt.Errorf("uknown user var: %s", n)
	}

	return result, nil
}

func templateTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}

func traverseMapStrings(n string, v reflect.Value, f traverseFunc) {
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

func traverseSliceStrings(n string, v reflect.Value, f traverseFunc) {
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

func traverseStructStrings(n string, v reflect.Value, f traverseFunc) {
	if n != "" {
		n = n + "."
	}

	vt := v.Type()
	for i := 0; i < vt.NumField(); i++ {
		field := v.FieldByIndex([]int{i})
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
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

func traverseValue(n string, v reflect.Value, f traverseFunc) (string, bool) {
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
