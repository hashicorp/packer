package function

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeTemplateDirFunc constructs a function that returns the directory
// in which the configuration file is located.
func MakeTemplateDirFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(baseDir), nil
		},
	})
}
