package common

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"reflect"
	"strconv"
	"text/template"
	"time"
)

// ConfigTemplate processes your entire configuration struct and processes
// all strings through the Golang text/template processor. This exposes
// common functions to all strings within Packer without any extra effort
// by the implementor.
type ConfigTemplate struct {
	BuilderVars map[string]string
	UserVars    map[string]string
	processed   map[string]struct{}
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
		processed:   make(map[string]struct{}),
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
	errs := make([]error, 0)

	f := func(n string, s string) string {
		if _, ok := ct.processed[n]; ok {
			return s
		}

		result, err := ct.processSingle(n)
		if err != nil {
			errs = append(errs, err)
		}

		return result
	}

	traverseStructStrings("", ct.v, f)
	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

// ProcessSingle processes a single element of configuration. If the
// configuration key has already been processed, it is an error. Once processed,
// this key will be skipped when Process is called.
func (ct *ConfigTemplate) ProcessSingle(n string) error {
	var err error
	found := false

	f := func(curN string, s string) string {
		if curN != n {
			return s
		}

		found = true

		var result string
		result, err = ct.processSingle(n)
		if err != nil {
			return s
		}

		return result
	}

	traverseStructStrings("", ct.v, f)
	if !found {
		err = fmt.Errorf("key '%s' not found", n)
	}

	return err
}

func (ct *ConfigTemplate) processSingle(n string) (string, error) {
	if _, ok := ct.processed[n]; ok {
		return "", fmt.Errorf("key already processed: %s", n)
	}

	t, ok := ct.t[n]
	if !ok {
		return "", fmt.Errorf("template not found: " + n)
	}

	buf := new(bytes.Buffer)
	err := t.Execute(buf, nil)
	if err != nil {
		return "", fmt.Errorf("Error processing %s: %s", n, err)
	}

	ct.processed[n] = struct{}{}
	return buf.String(), nil
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
