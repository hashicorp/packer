package function

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var Filebase64 = function.New(&function.Spec{
	Params: []function.Parameter{
		function.Parameter{
			Name:        "path",
			Description: "Read a file and encode it as a base64 string",
			Type:        cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		path := args[0].AsString()
		content, err := os.ReadFile(path)
		if err != nil {
			return cty.NullVal(cty.String), fmt.Errorf("failed to read file %q: %s", path, err)
		}

		out := &strings.Builder{}
		enc := base64.NewEncoder(base64.StdEncoding, out)
		_, err = enc.Write(content)
		if err != nil {
			return cty.NullVal(cty.String), fmt.Errorf("failed to write file %q as base64: %s", path, err)
		}
		_ = enc.Close()

		return cty.StringVal(out.String()), nil
	},
})
