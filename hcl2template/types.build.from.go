package hcl2template

import (
	"strings"
)

func sourceRefFromString(in string) SourceRef {
	args := strings.Split(in, ".")
	if len(args) < 2 {
		return NoSource
	}
	if len(args) > 2 {
		// source.type.name
		args = args[1:]
	}
	return SourceRef{
		Type: args[0],
		Name: args[1],
	}
}
