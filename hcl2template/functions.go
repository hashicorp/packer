package hcl2template

import (
	"fmt"

	"github.com/hashicorp/go-cty-funcs/cidr"
	"github.com/hashicorp/go-cty-funcs/collection"
	"github.com/hashicorp/go-cty-funcs/crypto"
	"github.com/hashicorp/go-cty-funcs/encoding"
	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/hashicorp/go-cty-funcs/uuid"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	pkrfunction "github.com/hashicorp/packer/hcl2template/function"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// Functions returns the set of functions that should be used to when
// evaluating expressions in the receiving scope.
//
// basedir is used with file functions and allows a user to reference a file
// using local path. Usually basedir is the directory in which the config file
// is located
//
func Functions(basedir string) map[string]function.Function {

	funcs := map[string]function.Function{
		"abs":                stdlib.AbsoluteFunc,
		"abspath":            filesystem.AbsPathFunc,
		"aws_secretsmanager": pkrfunction.AWSSecret,
		"basename":           filesystem.BasenameFunc,
		"base64decode":       encoding.Base64DecodeFunc,
		"base64encode":       encoding.Base64EncodeFunc,
		"bcrypt":             crypto.BcryptFunc,
		"can":                tryfunc.CanFunc,
		"ceil":               stdlib.CeilFunc,
		"chomp":              stdlib.ChompFunc,
		"chunklist":          stdlib.ChunklistFunc,
		"cidrhost":           cidr.HostFunc,
		"cidrnetmask":        cidr.NetmaskFunc,
		"cidrsubnet":         cidr.SubnetFunc,
		"cidrsubnets":        cidr.SubnetsFunc,
		"coalesce":           collection.CoalesceFunc,
		"coalescelist":       stdlib.CoalesceListFunc,
		"compact":            stdlib.CompactFunc,
		"concat":             stdlib.ConcatFunc,
		"consul_key":         pkrfunction.ConsulFunc,
		"contains":           stdlib.ContainsFunc,
		"convert":            typeexpr.ConvertFunc,
		"csvdecode":          stdlib.CSVDecodeFunc,
		"dirname":            filesystem.DirnameFunc,
		"distinct":           stdlib.DistinctFunc,
		"element":            stdlib.ElementFunc,
		"file":               filesystem.MakeFileFunc(basedir, false),
		"fileexists":         filesystem.MakeFileExistsFunc(basedir),
		"fileset":            filesystem.MakeFileSetFunc(basedir),
		"flatten":            stdlib.FlattenFunc,
		"floor":              stdlib.FloorFunc,
		"format":             stdlib.FormatFunc,
		"formatdate":         stdlib.FormatDateFunc,
		"formatlist":         stdlib.FormatListFunc,
		"indent":             stdlib.IndentFunc,
		"index":              stdlib.IndexFunc,
		"join":               stdlib.JoinFunc,
		"jsondecode":         stdlib.JSONDecodeFunc,
		"jsonencode":         stdlib.JSONEncodeFunc,
		"keys":               stdlib.KeysFunc,
		"length":             pkrfunction.LengthFunc,
		"log":                stdlib.LogFunc,
		"lookup":             stdlib.LookupFunc,
		"lower":              stdlib.LowerFunc,
		"max":                stdlib.MaxFunc,
		"md5":                crypto.Md5Func,
		"merge":              stdlib.MergeFunc,
		"min":                stdlib.MinFunc,
		"parseint":           stdlib.ParseIntFunc,
		"pathexpand":         filesystem.PathExpandFunc,
		"pow":                stdlib.PowFunc,
		"range":              stdlib.RangeFunc,
		"reverse":            stdlib.ReverseListFunc,
		"replace":            stdlib.ReplaceFunc,
		"regex_replace":      stdlib.RegexReplaceFunc,
		"rsadecrypt":         crypto.RsaDecryptFunc,
		"setintersection":    stdlib.SetIntersectionFunc,
		"setproduct":         stdlib.SetProductFunc,
		"setunion":           stdlib.SetUnionFunc,
		"sha1":               crypto.Sha1Func,
		"sha256":             crypto.Sha256Func,
		"sha512":             crypto.Sha512Func,
		"signum":             stdlib.SignumFunc,
		"slice":              stdlib.SliceFunc,
		"sort":               stdlib.SortFunc,
		"split":              stdlib.SplitFunc,
		"strrev":             stdlib.ReverseFunc,
		"substr":             stdlib.SubstrFunc,
		"timestamp":          pkrfunction.TimestampFunc,
		"timeadd":            stdlib.TimeAddFunc,
		"title":              stdlib.TitleFunc,
		"trim":               stdlib.TrimFunc,
		"trimprefix":         stdlib.TrimPrefixFunc,
		"trimspace":          stdlib.TrimSpaceFunc,
		"trimsuffix":         stdlib.TrimSuffixFunc,
		"try":                tryfunc.TryFunc,
		"upper":              stdlib.UpperFunc,
		"urlencode":          encoding.URLEncodeFunc,
		"uuidv4":             uuid.V4Func,
		"uuidv5":             uuid.V5Func,
		"values":             stdlib.ValuesFunc,
		"vault":              pkrfunction.VaultFunc,
		"yamldecode":         ctyyaml.YAMLDecodeFunc,
		"yamlencode":         ctyyaml.YAMLEncodeFunc,
		"zipmap":             stdlib.ZipmapFunc,
	}

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
