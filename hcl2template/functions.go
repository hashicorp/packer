package hcl2template

import (
	"fmt"

	"github.com/hashicorp/go-cty-funcs/cidr"
	"github.com/hashicorp/go-cty-funcs/crypto"
	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	pkrfunction "github.com/hashicorp/packer/hcl2template/function"
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
	var basedir string
	funcs := map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          filesystem.AbsPathFunc,
		"basename":         filesystem.BasenameFunc,
		"base64decode":     unimplFunc, // funcs.Base64DecodeFunc,
		"base64encode":     unimplFunc, // funcs.Base64EncodeFunc,
		"base64gzip":       unimplFunc, // funcs.Base64GzipFunc,
		"base64sha256":     crypto.Base64Sha256Func,
		"base64sha512":     crypto.Base64Sha512Func,
		"bcrypt":           crypto.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             unimplFunc, // funcs.CeilFunc,
		"chomp":            unimplFunc, // funcs.ChompFunc,
		"cidrhost":         cidr.HostFunc,
		"cidrnetmask":      cidr.NetmaskFunc,
		"cidrsubnet":       cidr.SubnetFunc,
		"cidrsubnets":      cidr.SubnetsFunc,
		"coalesce":         unimplFunc, // funcs.CoalesceFunc,
		"coalescelist":     unimplFunc, // funcs.CoalesceListFunc,
		"compact":          unimplFunc, // funcs.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         unimplFunc, // funcs.ContainsFunc,
		"convert":          typeexpr.ConvertFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          filesystem.DirnameFunc,
		"distinct":         unimplFunc, // funcs.DistinctFunc,
		"element":          unimplFunc, // funcs.ElementFunc,
		"chunklist":        unimplFunc, // funcs.ChunklistFunc,
		"file":             unimplFunc, // filesystem.MakeFileFunc(basedir, false),
		"fileexists":       filesystem.MakeFileExistsFunc(basedir),
		"fileset":          filesystem.MakeFileSetFunc(basedir),
		"filebase64":       unimplFunc, // funcs.MakeFileFunc(basedir, true),
		"filebase64sha256": crypto.MakeFileBase64Sha256Func(basedir),
		"filebase64sha512": crypto.MakeFileBase64Sha512Func(basedir),
		"filemd5":          crypto.MakeFileMd5Func(basedir),
		"filesha1":         crypto.MakeFileSha1Func(basedir),
		"filesha256":       crypto.MakeFileSha256Func(basedir),
		"filesha512":       crypto.MakeFileSha512Func(basedir),
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
		"md5":              crypto.Md5Func,
		"merge":            unimplFunc, // funcs.MergeFunc,
		"min":              stdlib.MinFunc,
		"parseint":         unimplFunc, // funcs.ParseIntFunc,
		"pathexpand":       filesystem.PathExpandFunc,
		"pow":              unimplFunc, // funcs.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          unimplFunc, // funcs.ReplaceFunc,
		"reverse":          unimplFunc, // funcs.ReverseFunc,
		"rsadecrypt":       crypto.RsaDecryptFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       unimplFunc, // funcs.SetProductFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             crypto.Sha1Func,
		"sha256":           crypto.Sha256Func,
		"sha512":           crypto.Sha512Func,
		"signum":           unimplFunc, // funcs.SignumFunc,
		"slice":            unimplFunc, // funcs.SliceFunc,
		"sort":             unimplFunc, // funcs.SortFunc,
		"split":            unimplFunc, // funcs.SplitFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"timestamp":        pkrfunction.TimestampFunc,
		"timeadd":          unimplFunc, // funcs.TimeAddFunc,
		"title":            unimplFunc, // funcs.TitleFunc,
		"trim":             unimplFunc, // funcs.TrimFunc,
		"trimprefix":       unimplFunc, // funcs.TrimPrefixFunc,
		"trimspace":        unimplFunc, // funcs.TrimSpaceFunc,
		"trimsuffix":       unimplFunc, // funcs.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        unimplFunc, // funcs.URLEncodeFunc,
		"uuid":             crypto.UUIDFunc,
		"uuidv5":           crypto.UUIDV5Func,
		"values":           unimplFunc, // funcs.ValuesFunc,
		"yamldecode":       ctyyaml.YAMLDecodeFunc,
		"yamlencode":       ctyyaml.YAMLEncodeFunc,
		"zipmap":           unimplFunc, // funcs.ZipmapFunc,
	}

	// s.funcs["templatefile"] = funcs.MakeTemplateFileFunc(basedir, func() map[string]function.Function {
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
