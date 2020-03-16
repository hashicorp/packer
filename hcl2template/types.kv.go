//go:generate mapstructure-to-hcl2 -type KeyValue,KeyValues,KVFilter

package hcl2template

type KeyValue struct {
	Key   string
	Value string
}

type KeyValues []KeyValue

func (kvs KeyValues) CopyOn(to *map[string]string) []error {
	if *to == nil {
		*to = map[string]string{}
	}
	for _, kv := range kvs {
		(*to)[kv.Key] = kv.Value
	}
	return nil
}

type KVFilter struct {
	Filters map[string]string
	Filter  KeyValues
}

func (kvf *KVFilter) Prepare() []error {
	kvf.Filter.CopyOn(&kvf.Filters)
	return nil
}

func (kvf *KVFilter) Empty() bool {
	return len(kvf.Filters) == 0
}
