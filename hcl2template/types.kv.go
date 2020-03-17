//go:generate mapstructure-to-hcl2 -type NameValue,NameValues,KVFilter

package hcl2template

type NameValue struct {
	Name  string
	Value string
}

type NameValues []NameValue

func (kvs NameValues) CopyOn(to *map[string]string) []error {
	if len(kvs) == 0 {
		return nil
	}
	if *to == nil {
		*to = map[string]string{}
	}
	for _, kv := range kvs {
		(*to)[kv.Name] = kv.Value
	}
	return nil
}

type KVFilter struct {
	Filters map[string]string
	Filter  NameValues
}

func (kvf *KVFilter) Prepare() []error {
	kvf.Filter.CopyOn(&kvf.Filters)
	return nil
}

func (kvf *KVFilter) Empty() bool {
	return len(kvf.Filters) == 0
}
