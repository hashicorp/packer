//go:generate mapstructure-to-hcl2 -type KeyValue

package hcl2template

type KeyValue struct {
	Key   string
	Value string
}

type KVFilter struct {
	Filters map[string]string
	Filter  []KeyValue
}

func (kvf *KVFilter) Prepare() error {
	for _, filter := range kvf.Filter {
		kvf.Filters[filter.Key] = filter.Value
	}
	return nil
}

func (kvf *KVFilter) Empty() bool {
	return len(kvf.Filters) == 0
}
