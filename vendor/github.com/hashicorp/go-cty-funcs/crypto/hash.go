package crypto

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// Base64Sha256Func is a function that computes the SHA256 hash of a given
// string and encodes it with Base64.
var Base64Sha256Func = makeStringHashFunction(sha256.New, base64.StdEncoding.EncodeToString)

// MakeFileBase64Sha256Func is a function that is like Base64Sha256Func but
// reads the contents of a file rather than hashing a given literal string.
func MakeFileBase64Sha256Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, sha256.New, base64.StdEncoding.EncodeToString)
}

// Base64Sha512Func is a function that computes the SHA256 hash of a given
// string and encodes it with Base64.
var Base64Sha512Func = makeStringHashFunction(sha512.New, base64.StdEncoding.EncodeToString)

// MakeFileBase64Sha512Func is a function that is like Base64Sha512Func but
// reads the contents of a file rather than hashing a given literal string.
func MakeFileBase64Sha512Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, sha512.New, base64.StdEncoding.EncodeToString)
}

// Md5Func is a function that computes the MD5 hash of a given string and
// encodes it with hexadecimal digits.
var Md5Func = makeStringHashFunction(md5.New, hex.EncodeToString)

// MakeFileMd5Func is a function that is like Md5Func but reads the contents of
// a file rather than hashing a given literal string.
func MakeFileMd5Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, md5.New, hex.EncodeToString)
}

// RsaDecryptFunc is a function that decrypts an RSA-encrypted ciphertext.
var RsaDecryptFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "ciphertext",
			Type: cty.String,
		},
		{
			Name: "privatekey",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		s := args[0].AsString()
		key := args[1].AsString()

		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("failed to decode input %q: cipher text must be base64-encoded", s)
		}

		block, _ := pem.Decode([]byte(key))
		if block == nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("failed to parse key: no key found")
		}
		if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
			return cty.UnknownVal(cty.String), fmt.Errorf(
				"failed to parse key: password protected keys are not supported. Please decrypt the key prior to use",
			)
		}

		x509Key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		out, err := rsa.DecryptPKCS1v15(nil, x509Key, b)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		return cty.StringVal(string(out)), nil
	},
})

// Sha1Func is a function that computes the SHA1 hash of a given string and
// encodes it with hexadecimal digits.
var Sha1Func = makeStringHashFunction(sha1.New, hex.EncodeToString)

// MakeFileSha1Func is a function that is like Sha1Func but reads the contents
// of a file rather than hashing a given literal string.
func MakeFileSha1Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, sha1.New, hex.EncodeToString)
}

// Sha256Func is a function that computes the SHA256 hash of a given string and
// encodes it with hexadecimal digits.
var Sha256Func = makeStringHashFunction(sha256.New, hex.EncodeToString)

// MakeFileSha256Func is a function that is like Sha256Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileSha256Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, sha256.New, hex.EncodeToString)
}

// Sha512Func is a function that computes the SHA512 hash of a given string and
// encodes it with hexadecimal digits.
var Sha512Func = makeStringHashFunction(sha512.New, hex.EncodeToString)

// MakeFileSha512Func is a function that is like Sha512Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileSha512Func(baseDir string) function.Function {
	return makeFileHashFunction(baseDir, sha512.New, hex.EncodeToString)
}

func makeStringHashFunction(hf func() hash.Hash, enc func([]byte) string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "str",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			s := args[0].AsString()
			h := hf()
			h.Write([]byte(s))
			rv := enc(h.Sum(nil))
			return cty.StringVal(rv), nil
		},
	})
}

func makeFileHashFunction(baseDir string, hf func() hash.Hash, enc func([]byte) string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			path := args[0].AsString()
			src, err := readFileBytes(baseDir, path)
			if err != nil {
				return cty.UnknownVal(cty.String), err
			}

			h := hf()
			h.Write(src)
			rv := enc(h.Sum(nil))
			return cty.StringVal(rv), nil
		},
	})
}

// Base64Sha256 computes the SHA256 hash of a given string and encodes it with
// Base64.
//
// The given string is first encoded as UTF-8 and then the SHA256 algorithm is
// applied as defined in RFC 4634. The raw hash is then encoded with Base64
// before returning. Terraform uses the "standard" Base64 alphabet as defined
// in RFC 4648 section 4.
func Base64Sha256(str cty.Value) (cty.Value, error) {
	return Base64Sha256Func.Call([]cty.Value{str})
}

// Base64Sha512 computes the SHA512 hash of a given string and encodes it with
// Base64.
//
// The given string is first encoded as UTF-8 and then the SHA256 algorithm is
// applied as defined in RFC 4634. The raw hash is then encoded with Base64
// before returning. Terraform uses the "standard" Base64  alphabet as defined
// in RFC 4648 section 4
func Base64Sha512(str cty.Value) (cty.Value, error) {
	return Base64Sha512Func.Call([]cty.Value{str})
}

// Md5 computes the MD5 hash of a given string and encodes it with hexadecimal
// digits.
func Md5(str cty.Value) (cty.Value, error) {
	return Md5Func.Call([]cty.Value{str})
}

// RsaDecrypt decrypts an RSA-encrypted ciphertext, returning the corresponding
// cleartext.
func RsaDecrypt(ciphertext, privatekey cty.Value) (cty.Value, error) {
	return RsaDecryptFunc.Call([]cty.Value{ciphertext, privatekey})
}

// Sha1 computes the SHA1 hash of a given string and encodes it with
// hexadecimal digits.
func Sha1(str cty.Value) (cty.Value, error) {
	return Sha1Func.Call([]cty.Value{str})
}

// Sha256 computes the SHA256 hash of a given string and encodes it with
// hexadecimal digits.
func Sha256(str cty.Value) (cty.Value, error) {
	return Sha256Func.Call([]cty.Value{str})
}

// Sha512 computes the SHA512 hash of a given string and encodes it with
// hexadecimal digits.
func Sha512(str cty.Value) (cty.Value, error) {
	return Sha512Func.Call([]cty.Value{str})
}

func readFileBytes(baseDir, path string) ([]byte, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand ~: %s", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	// Ensure that the path is canonical for the host OS
	path = filepath.Clean(path)

	src, err := ioutil.ReadFile(path)
	if err != nil {
		// ReadFile does not return Terraform-user-friendly error messages, so
		// we'll provide our own.
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no file exists at %s", path)
		}
		return nil, fmt.Errorf("failed to read %s", path)
	}

	return src, nil
}
