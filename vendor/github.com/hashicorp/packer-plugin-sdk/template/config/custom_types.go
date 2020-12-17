//go:generate mapstructure-to-hcl2 -type KeyValue,KeyValues,KeyValueFilter,NameValue,NameValues,NameValueFilter
package config

import (
	"strconv"
)

type Trilean uint8

const (
	// This will assign unset to 0, which is the default value in interpolation
	TriUnset Trilean = iota
	TriTrue
	TriFalse
)

func (t Trilean) ToString() string {
	if t == TriTrue {
		return "TriTrue"
	} else if t == TriFalse {
		return "TriFalse"
	}
	return "TriUnset"
}

func (t Trilean) ToBoolPointer() *bool {
	if t == TriTrue {
		return boolPointer(true)
	} else if t == TriFalse {
		return boolPointer(false)
	}
	return nil
}

func (t Trilean) True() bool {
	if t == TriTrue {
		return true
	}
	return false
}

func (t Trilean) False() bool {
	if t == TriFalse {
		return true
	}
	return false
}

func TrileanFromString(s string) (Trilean, error) {
	if s == "" {
		return TriUnset, nil
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return TriUnset, err
	} else if b == true {
		return TriTrue, nil
	} else {
		return TriFalse, nil
	}
}

func TrileanFromBool(b bool) Trilean {
	if b {
		return TriTrue
	}
	return TriFalse
}

func boolPointer(b bool) *bool {
	return &b
}

// These are used to convert HCL blocks to key-value pairs
type KeyValue struct {
	Key   string
	Value string
}

type KeyValues []KeyValue

func (kvs KeyValues) CopyOn(to *map[string]string) []error {
	if len(kvs) == 0 {
		return nil
	}
	if *to == nil {
		*to = map[string]string{}
	}
	for _, kv := range kvs {
		(*to)[kv.Key] = kv.Value
	}
	return nil
}

type KeyValueFilter struct {
	Filters map[string]string
	Filter  KeyValues
}

func (kvf *KeyValueFilter) Prepare() []error {
	kvf.Filter.CopyOn(&kvf.Filters)
	return nil
}

func (kvf *KeyValueFilter) Empty() bool {
	return len(kvf.Filters) == 0
}

type NameValue struct {
	Name  string
	Value string
}

type NameValues []NameValue

func (nvs NameValues) CopyOn(to *map[string]string) []error {
	if len(nvs) == 0 {
		return nil
	}
	if *to == nil {
		*to = map[string]string{}
	}
	for _, kv := range nvs {
		(*to)[kv.Name] = kv.Value
	}
	return nil
}

type NameValueFilter struct {
	Filters map[string]string
	Filter  NameValues
}

func (nvf *NameValueFilter) Prepare() []error {
	nvf.Filter.CopyOn(&nvf.Filters)
	return nil
}

func (nvf *NameValueFilter) Empty() bool {
	return len(nvf.Filters) == 0
}
