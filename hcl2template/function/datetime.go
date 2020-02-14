package function

import (
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// TimestampFunc constructs a function that returns a string representation of the current date and time.
var TimestampFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(time.Now().UTC().Format(time.RFC3339)), nil
	},
})

// Timestamp returns a string representation of the current date and time.
//
// In the HCL language, timestamps are conventionally represented as strings
// using RFC 3339 "Date and Time format" syntax, and so timestamp returns a
// string in this format.
func Timestamp() (cty.Value, error) {
	return TimestampFunc.Call([]cty.Value{})
}
