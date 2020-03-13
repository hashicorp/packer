//go:generate mapstructure-to-hcl2 -type KeyValue

package hcl2template

type KeyValue struct {
	Key   string
	Value string
}
