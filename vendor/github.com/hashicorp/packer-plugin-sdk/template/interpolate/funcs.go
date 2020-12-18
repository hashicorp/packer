package interpolate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	commontpl "github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	strftime "github.com/jehiah/go-strftime"
)

// InitTime is the UTC time when this package was initialized. It is
// used as the timestamp for all configuration templates so that they
// match for a single build.
var InitTime time.Time

func init() {
	InitTime = time.Now().UTC()
}

// Funcs are the interpolation funcs that are available within interpolations.
var FuncGens = map[string]interface{}{
	"build_name":         funcGenBuildName,
	"build_type":         funcGenBuildType,
	"env":                funcGenEnv,
	"isotime":            funcGenIsotime,
	"strftime":           funcGenStrftime,
	"pwd":                funcGenPwd,
	"split":              funcGenSplitter,
	"template_dir":       funcGenTemplateDir,
	"timestamp":          funcGenTimestamp,
	"uuid":               funcGenUuid,
	"user":               funcGenUser,
	"packer_version":     funcGenPackerVersion,
	"consul_key":         funcGenConsul,
	"vault":              funcGenVault,
	"sed":                funcGenSed,
	"build":              funcGenBuild,
	"aws_secretsmanager": funcGenAwsSecrets,

	"replace":     replace,
	"replace_all": replace_all,

	"upper": strings.ToUpper,
	"lower": strings.ToLower,
}

var ErrVariableNotSetString = "Error: variable not set:"

// FuncGenerator is a function that given a context generates a template
// function for the template.
type FuncGenerator func(*Context) interface{}

// Funcs returns the functions that can be used for interpolation given
// a context.
func Funcs(ctx *Context) template.FuncMap {
	result := make(map[string]interface{})
	for k, v := range FuncGens {
		switch v := v.(type) {
		case func(*Context) interface{}:
			result[k] = v(ctx)
		default:
			result[k] = v
		}
	}
	if ctx != nil {
		for k, v := range ctx.Funcs {
			result[k] = v
		}
	}

	return template.FuncMap(result)
}

func funcGenSplitter(ctx *Context) interface{} {
	return func(k string, s string, i int) (string, error) {
		// return func(s string) (string, error) {
		split := strings.Split(k, s)
		if len(split) <= i {
			return "", fmt.Errorf("the substring %d was unavailable using the separator value, %s, only %d values were found", i, s, len(split))
		}
		return split[i], nil
	}
}

func funcGenBuildName(ctx *Context) interface{} {
	return func() (string, error) {
		if ctx == nil || ctx.BuildName == "" {
			return "", errors.New("build_name not available")
		}

		return ctx.BuildName, nil
	}
}

func funcGenBuildType(ctx *Context) interface{} {
	return func() (string, error) {
		if ctx == nil || ctx.BuildType == "" {
			return "", errors.New("build_type not available")
		}

		return ctx.BuildType, nil
	}
}

func funcGenEnv(ctx *Context) interface{} {
	return func(k string) (string, error) {
		if !ctx.EnableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("env vars are not allowed here")
		}

		return os.Getenv(k), nil
	}
}

func funcGenIsotime(ctx *Context) interface{} {
	return func(format ...string) (string, error) {
		if len(format) == 0 {
			return InitTime.Format(time.RFC3339), nil
		}

		if len(format) > 1 {
			return "", fmt.Errorf("too many values, 1 needed: %v", format)
		}

		return InitTime.Format(format[0]), nil
	}
}

func funcGenStrftime(ctx *Context) interface{} {
	return func(format string) string {
		return strftime.Format(format, InitTime)
	}
}

func funcGenPwd(ctx *Context) interface{} {
	return func() (string, error) {
		return os.Getwd()
	}
}

func funcGenTemplateDir(ctx *Context) interface{} {
	return func() (string, error) {
		if ctx == nil || ctx.TemplatePath == "" {
			return "", errors.New("template path not available")
		}

		path, err := filepath.Abs(filepath.Dir(ctx.TemplatePath))
		if err != nil {
			return "", err
		}

		return path, nil
	}
}

func passthroughOrInterpolate(data map[interface{}]interface{}, s string) (string, error) {
	if heldPlace, ok := data[s]; ok {
		if hp, ok := heldPlace.(string); ok {
			// If we're in the first interpolation pass, the goal is to
			// make sure that we pass the value through.
			// TODO match against an actual string constant
			if strings.Contains(hp, packerbuilderdata.PlaceholderMsg) {
				return fmt.Sprintf("{{.%s}}", s), nil
			} else {
				return hp, nil
			}
		}
	}
	return "", fmt.Errorf("loaded data, but couldnt find %s in it.", s)

}
func funcGenBuild(ctx *Context) interface{} {
	// Depending on where the context data is coming from, it could take a few
	// different map types. The following switch standardizes the map types
	// so we can act on them correctly.
	return func(s string) (string, error) {
		switch data := ctx.Data.(type) {
		case map[interface{}]interface{}:
			return passthroughOrInterpolate(data, s)
		case map[string]interface{}:
			// convert to a map[interface{}]interface{} so we can use same
			// parsing on it
			passed := make(map[interface{}]interface{}, len(data))
			for k, v := range data {
				passed[k] = v
			}
			return passthroughOrInterpolate(passed, s)
		case map[string]string:
			// convert to a map[interface{}]interface{} so we can use same
			// parsing on it
			passed := make(map[interface{}]interface{}, len(data))
			for k, v := range data {
				passed[k] = v
			}
			return passthroughOrInterpolate(passed, s)
		default:
			return "", fmt.Errorf("Error validating build variable: the given "+
				"variable %s will not be passed into your plugin.", s)
		}
	}
}

func funcGenTimestamp(ctx *Context) interface{} {
	return func() string {
		return strconv.FormatInt(InitTime.Unix(), 10)
	}
}

func funcGenUser(ctx *Context) interface{} {
	return func(k string) (string, error) {
		if ctx == nil || ctx.UserVariables == nil {
			return "", errors.New("test")
		}

		val, ok := ctx.UserVariables[k]
		if ctx.EnableEnv {
			// error and retry if we're interpolating UserVariables. But if
			// we're elsewhere in the template, just return the empty string.
			if !ok {
				return "", fmt.Errorf("%s %s", ErrVariableNotSetString, k)
			}
		}
		return val, nil
	}
}

func funcGenUuid(ctx *Context) interface{} {
	return func() string {
		return uuid.TimeOrderedUUID()
	}
}

func funcGenPackerVersion(ctx *Context) interface{} {
	return func() (string, error) {
		if ctx == nil || ctx.CorePackerVersionString == "" {
			return "", errors.New("packer_version not available")
		}

		return ctx.CorePackerVersionString, nil
	}
}

func funcGenConsul(ctx *Context) interface{} {
	return func(key string) (string, error) {
		if !ctx.EnableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("consul_key is not allowed here")
		}

		return commontpl.Consul(key)
	}
}

func funcGenVault(ctx *Context) interface{} {
	return func(path string, key string) (string, error) {
		// Only allow interpolation from Vault when env vars are being read.
		if !ctx.EnableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("Vault vars are only allowed in the variables section")
		}

		return commontpl.Vault(path, key)
	}
}

func funcGenAwsSecrets(ctx *Context) interface{} {
	return func(secret ...string) (string, error) {
		if !ctx.EnableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("AWS Secrets Manager is only allowed in the variables section")
		}
		switch len(secret) {
		case 0:
			return "", errors.New("secret name must be provided")
		case 1:
			return commontpl.GetAWSSecret(secret[0], "")
		case 2:
			return commontpl.GetAWSSecret(secret[0], secret[1])
		default:
			return "", errors.New("only secret name and optional secret key can be provided.")
		}
	}
}

func funcGenSed(ctx *Context) interface{} {
	return func(expression string, inputString string) (string, error) {
		return "", errors.New("template function `sed` is deprecated " +
			"use `replace` or `replace_all` instead." +
			"Documentation: https://www.packer.io/docs/templates/engine")
	}
}

func replace_all(old, new, src string) string {
	return strings.ReplaceAll(src, old, new)
}

func replace(old, new string, n int, src string) string {
	return strings.Replace(src, old, new, n)
}
