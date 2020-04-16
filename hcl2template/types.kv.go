//go:generate mapstructure-to-hcl2 -type KeyValue,KeyValues,KeyValueFilter,NameValue,NameValues,NameValueFilter

package hcl2template

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
