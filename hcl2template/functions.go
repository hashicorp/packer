// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"bytes"
	"encoding/base64"
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
	"golang.org/x/text/encoding/ianaindex"
)

// Functions returns the set of functions that should be used to when
// evaluating expressions in the receiving scope.
//
// basedir is used with file functions and allows a user to reference a file
// using local path. Usually basedir is the directory in which the config file
// is located
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
		"index":              pkrfunction.IndexFunc, // stdlib.IndexFunc is not compatible
		"join":               stdlib.JoinFunc,
		"jsondecode":         stdlib.JSONDecodeFunc,
		"jsonencode":         stdlib.JSONEncodeFunc,
		"keys":               stdlib.KeysFunc,
		"legacy_isotime":     pkrfunction.LegacyIsotimeFunc,
		"legacy_strftime":    pkrfunction.LegacyStrftimeFunc,
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
		"regex":              stdlib.RegexFunc,
		"regexall":           stdlib.RegexAllFunc,
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
		"textdecodebase64":   TextDecodeBase64Func,
		"textencodebase64":   TextEncodeBase64Func,
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

	funcs["templatefile"] = pkrfunction.MakeTemplateFileFunc(basedir, func() map[string]function.Function {
		// The templatefile function prevents recursive calls to itself
		// by copying this map and overwriting the "templatefile" entry.
		return funcs
	})

	return funcs
}

// TextEncodeBase64Func constructs a function that encodes a string to a target encoding and then to a base64 sequence.
var TextEncodeBase64Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "string",
			Type: cty.String,
		},
		{
			Name: "encoding",
			Type: cty.String,
		},
	},
	Description:  "Encodes the input string (UTF-8) to the destination encoding. The output is base64 to account for cty limiting strings to NFC normalised UTF-8 strings.",
	Type:         function.StaticReturnType(cty.String),
	RefineResult: func(rb *cty.RefinementBuilder) *cty.RefinementBuilder { return rb.NotNull() },
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		encoding, err := ianaindex.IANA.Encoding(args[1].AsString())
		if err != nil || encoding == nil {
			return cty.UnknownVal(cty.String), function.NewArgErrorf(1, "%q is not a supported IANA encoding name or alias", args[1].AsString())
		}

		encName, err := ianaindex.IANA.Name(encoding)
		if err != nil { // would be weird, since we just read this encoding out
			encName = args[1].AsString()
		}

		encoder := encoding.NewEncoder()
		encodedInput, err := encoder.Bytes([]byte(args[0].AsString()))
		if err != nil {
			// The string representations of "err" disclose implementation
			// details of the underlying library, and the main error we might
			// like to return a special message for is unexported as
			// golang.org/x/text/encoding/internal.RepertoireError, so this
			// is just a generic error message for now.
			//
			// We also don't include the string itself in the message because
			// it can typically be very large, contain newline characters,
			// etc.
			return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given string contains characters that cannot be represented in %s", encName)
		}

		return cty.StringVal(base64.StdEncoding.EncodeToString(encodedInput)), nil
	},
})

// TextDecodeBase64Func constructs a function that decodes a base64 sequence from the source encoding to UTF-8.
var TextDecodeBase64Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "source",
			Type: cty.String,
		},
		{
			Name: "encoding",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	Description:  "Encodes the input base64 blob from an encoding to utf-8. The input is base64 to account for cty limiting strings to NFC normalised UTF-8 strings.",
	RefineResult: func(rb *cty.RefinementBuilder) *cty.RefinementBuilder { return rb.NotNull() },
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		encoding, err := ianaindex.IANA.Encoding(args[1].AsString())
		if err != nil || encoding == nil {
			return cty.UnknownVal(cty.String), function.NewArgErrorf(1, "%q is not a supported IANA encoding name or alias", args[1].AsString())
		}

		encName, err := ianaindex.IANA.Name(encoding)
		if err != nil { // would be weird, since we just read this encoding out
			encName = args[1].AsString()
		}

		s := args[0].AsString()
		sDec, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			switch err := err.(type) {
			case base64.CorruptInputError:
				return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given value is has an invalid base64 symbol at offset %d", int(err))
			default:
				return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "invalid source string: %w", err)
			}
		}

		decoder := encoding.NewDecoder()
		decoded, err := decoder.Bytes(sDec)
		if err != nil || bytes.ContainsRune(decoded, '�') {
			return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given string contains symbols that are not defined for %s", encName)
		}

		return cty.StringVal(string(decoded)), nil
	},
})

var unimplFunc = function.New(&function.Spec{
	Type: func([]cty.Value) (cty.Type, error) {
		return cty.DynamicPseudoType, fmt.Errorf("function not yet implemented")
	},
	Impl: func([]cty.Value, cty.Type) (cty.Value, error) {
		return cty.DynamicVal, fmt.Errorf("function not yet implemented")
	},
})
