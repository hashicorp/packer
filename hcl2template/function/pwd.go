package function

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// MakePwdFunc constructs a function that returns the working directory as a string.
func MakePwdFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			dir, err := os.Getwd()
			return cty.StringVal(dir), err
		},
	})
}
