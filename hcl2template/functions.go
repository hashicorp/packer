package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// Functions returns the set of functions that should be used to when evaluating
// expressions in the receiving scope.
func Functions() map[string]function.Function {

	// Our functions are from the cty stdlib functions. A lot HLC2 funcs are
	// defined in "github.com/hashicorp/terraform/lang/funcs" for now we will
	// only use/import the stdlib funcs to later on copy the usefull ones to
	// the stdlib.

	funcs := map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          unimplFunc, // funcs.AbsPathFunc,
		"basename":         unimplFunc, // funcs.BasenameFunc,
		"base64decode":     unimplFunc, // funcs.Base64DecodeFunc,
		"base64encode":     unimplFunc, // funcs.Base64EncodeFunc,
		"base64gzip":       unimplFunc, // funcs.Base64GzipFunc,
		"base64sha256":     unimplFunc, // funcs.Base64Sha256Func,
		"base64sha512":     unimplFunc, // funcs.Base64Sha512Func,
		"bcrypt":           unimplFunc, // funcs.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             unimplFunc, // funcs.CeilFunc,
		"chomp":            unimplFunc, // funcs.ChompFunc,
		"cidrhost":         unimplFunc, // funcs.CidrHostFunc,
		"cidrnetmask":      unimplFunc, // funcs.CidrNetmaskFunc,
		"cidrsubnet":       unimplFunc, // funcs.CidrSubnetFunc,
		"cidrsubnets":      unimplFunc, // funcs.CidrSubnetsFunc,
		"coalesce":         unimplFunc, // funcs.CoalesceFunc,
		"coalescelist":     unimplFunc, // funcs.CoalesceListFunc,
		"compact":          unimplFunc, // funcs.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         unimplFunc, // funcs.ContainsFunc,
		"convert":          typeexpr.ConvertFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          unimplFunc, // funcs.DirnameFunc,
		"distinct":         unimplFunc, // funcs.DistinctFunc,
		"element":          unimplFunc, // funcs.ElementFunc,
		"chunklist":        unimplFunc, // funcs.ChunklistFunc,
		"file":             unimplFunc, // funcs.MakeFileFunc(s.BaseDir, false),
		"fileexists":       unimplFunc, // funcs.MakeFileExistsFunc(s.BaseDir),
		"fileset":          unimplFunc, // funcs.MakeFileSetFunc(s.BaseDir),
		"filebase64":       unimplFunc, // funcs.MakeFileFunc(s.BaseDir, true),
		"filebase64sha256": unimplFunc, // funcs.MakeFileBase64Sha256Func(s.BaseDir),
		"filebase64sha512": unimplFunc, // funcs.MakeFileBase64Sha512Func(s.BaseDir),
		"filemd5":          unimplFunc, // funcs.MakeFileMd5Func(s.BaseDir),
		"filesha1":         unimplFunc, // funcs.MakeFileSha1Func(s.BaseDir),
		"filesha256":       unimplFunc, // funcs.MakeFileSha256Func(s.BaseDir),
		"filesha512":       unimplFunc, // funcs.MakeFileSha512Func(s.BaseDir),
		"flatten":          unimplFunc, // funcs.FlattenFunc,
		"floor":            unimplFunc, // funcs.FloorFunc,
		"format":           stdlib.FormatFunc,
		"formatdate":       stdlib.FormatDateFunc,
		"formatlist":       stdlib.FormatListFunc,
		"indent":           unimplFunc, // funcs.IndentFunc,
		"index":            unimplFunc, // funcs.IndexFunc,
		"join":             unimplFunc, // funcs.JoinFunc,
		"jsondecode":       stdlib.JSONDecodeFunc,
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             unimplFunc, // funcs.KeysFunc,
		"length":           unimplFunc, // funcs.LengthFunc,
		"list":             unimplFunc, // funcs.ListFunc,
		"log":              unimplFunc, // funcs.LogFunc,
		"lookup":           unimplFunc, // funcs.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"map":              unimplFunc, // funcs.MapFunc,
		"matchkeys":        unimplFunc, // funcs.MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              unimplFunc, // funcs.Md5Func,
		"merge":            unimplFunc, // funcs.MergeFunc,
		"min":              stdlib.MinFunc,
		"parseint":         unimplFunc, // funcs.ParseIntFunc,
		"pathexpand":       unimplFunc, // funcs.PathExpandFunc,
		"pow":              unimplFunc, // funcs.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          unimplFunc, // funcs.ReplaceFunc,
		"reverse":          unimplFunc, // funcs.ReverseFunc,
		"rsadecrypt":       unimplFunc, // funcs.RsaDecryptFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       unimplFunc, // funcs.SetProductFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             unimplFunc, // funcs.Sha1Func,
		"sha256":           unimplFunc, // funcs.Sha256Func,
		"sha512":           unimplFunc, // funcs.Sha512Func,
		"signum":           unimplFunc, // funcs.SignumFunc,
		"slice":            unimplFunc, // funcs.SliceFunc,
		"sort":             unimplFunc, // funcs.SortFunc,
		"split":            unimplFunc, // funcs.SplitFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"timestamp":        unimplFunc, // funcs.TimestampFunc,
		"timeadd":          unimplFunc, // funcs.TimeAddFunc,
		"title":            unimplFunc, // funcs.TitleFunc,
		"trim":             unimplFunc, // funcs.TrimFunc,
		"trimprefix":       unimplFunc, // funcs.TrimPrefixFunc,
		"trimspace":        unimplFunc, // funcs.TrimSpaceFunc,
		"trimsuffix":       unimplFunc, // funcs.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        unimplFunc, // funcs.URLEncodeFunc,
		"uuid":             unimplFunc, // funcs.UUIDFunc,
		"uuidv5":           unimplFunc, // funcs.UUIDV5Func,
		"values":           unimplFunc, // funcs.ValuesFunc,
		"yamldecode":       ctyyaml.YAMLDecodeFunc,
		"yamlencode":       ctyyaml.YAMLEncodeFunc,
		"zipmap":           unimplFunc, // funcs.ZipmapFunc,
	}

	// s.funcs["templatefile"] = funcs.MakeTemplateFileFunc(s.BaseDir, func() map[string]function.Function {
	// 	// The templatefile function prevents recursive calls to itself
	// 	// by copying this map and overwriting the "templatefile" entry.
	// 	return s.funcs
	// })

	return funcs
}

var unimplFunc = function.New(&function.Spec{
	Type: func([]cty.Value) (cty.Type, error) {
		return cty.DynamicPseudoType, fmt.Errorf("function not yet implemented")
	},
	Impl: func([]cty.Value, cty.Type) (cty.Value, error) {
		return cty.DynamicVal, fmt.Errorf("function not yet implemented")
	},
})
