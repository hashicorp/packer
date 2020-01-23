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

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/version"
	vaultapi "github.com/hashicorp/vault/api"
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
	"build_name":     funcGenBuildName,
	"build_type":     funcGenBuildType,
	"env":            funcGenEnv,
	"isotime":        funcGenIsotime,
	"strftime":       funcGenStrftime,
	"pwd":            funcGenPwd,
	"split":          funcGenSplitter,
	"template_dir":   funcGenTemplateDir,
	"timestamp":      funcGenTimestamp,
	"uuid":           funcGenUuid,
	"user":           funcGenUser,
	"packer_version": funcGenPackerVersion,
	"consul_key":     funcGenConsul,
	"vault":          funcGenVault,
	"sed":            funcGenSed,
	"build":          funcGenBuild,

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

func funcGenBuild(ctx *Context) interface{} {
	return func(s string) (string, error) {
		if data, ok := ctx.Data.(map[string]string); ok {
			if heldPlace, ok := data[s]; ok {
				// If we're in the first interpolation pass, the goal is to
				// make sure that we pass the value through.
				// TODO match against an actual string constant
				if strings.Contains(heldPlace, common.PlaceholderMsg) {
					return fmt.Sprintf("{{.%s}}", s), nil
				} else {
					return heldPlace, nil
				}
			}
			return "", fmt.Errorf("loaded data, but couldnt find %s in it.", s)
		}
		if data, ok := ctx.Data.(map[interface{}]interface{}); ok {
			// PlaceholderData has been passed into generator, so if the given
			// key already exists in data, then we know it's an "allowed" key
			if heldPlace, ok := data[s]; ok {
				if hp, ok := heldPlace.(string); ok {
					// If we're in the first interpolation pass, the goal is to
					// make sure that we pass the value through.
					// TODO match against an actual string constant
					if strings.Contains(hp, common.PlaceholderMsg) {
						return fmt.Sprintf("{{.%s}}", s), nil
					} else {
						return hp, nil
					}
				}
			}
			return "", fmt.Errorf("loaded data, but couldnt find %s in it.", s)
		}

		return "", fmt.Errorf("Error validating build variable: the given "+
			"variable %s will not be passed into your plugin.", s)
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
	return func() string {
		return version.FormattedVersion()
	}
}

func funcGenConsul(ctx *Context) interface{} {
	return func(k string) (string, error) {
		if !ctx.EnableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("consul_key is not allowed here")
		}

		consulConfig := consulapi.DefaultConfig()
		client, err := consulapi.NewClient(consulConfig)
		if err != nil {
			return "", fmt.Errorf("error getting consul client: %s", err)
		}

		q := &consulapi.QueryOptions{}
		kv, _, err := client.KV().Get(k, q)
		if err != nil {
			return "", fmt.Errorf("error reading consul key: %s", err)
		}
		if kv == nil {
			return "", fmt.Errorf("key does not exist at the given path: %s", k)
		}

		value := string(kv.Value)
		if value == "" {
			return "", fmt.Errorf("value is empty at path %s", k)
		}

		return value, nil
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
		if token := os.Getenv("VAULT_TOKEN"); token == "" {
			return "", errors.New("Must set VAULT_TOKEN env var in order to " +
				"use vault template function")
		}
		// const EnvVaultAddress = "VAULT_ADDR"
		// const EnvVaultToken = "VAULT_TOKEN"
		vaultConfig := vaultapi.DefaultConfig()
		cli, err := vaultapi.NewClient(vaultConfig)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Error getting Vault client: %s", err))
		}
		secret, err := cli.Logical().Read(path)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Error reading vault secret: %s", err))
		}
		if secret == nil {
			return "", errors.New(fmt.Sprintf("Vault Secret does not exist at the given path."))
		}

		data, ok := secret.Data["data"]
		if !ok {
			// maybe ths is v1, not v2 kv store
			value, ok := secret.Data[key]
			if ok {
				return value.(string), nil
			}

			// neither v1 nor v2 proudced a valid value
			return "", errors.New(fmt.Sprintf("Vault data was empty at the "+
				"given path. Warnings: %s", strings.Join(secret.Warnings, "; ")))
		}

		value := data.(map[string]interface{})[key].(string)
		return value, nil
	}
}

func funcGenSed(ctx *Context) interface{} {
	return func(expression string, inputString string) (string, error) {
		return "", errors.New("template function `sed` is deprecated " +
			"use `replace` or `replace_all` instead." +
			"Documentation: https://www.packer.io/docs/templates/engine.html")
	}
}

func replace_all(old, new, src string) string {
	return strings.ReplaceAll(src, old, new)
}

func replace(old, new string, n int, src string) string {
	return strings.Replace(src, old, new, n)
}
