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
	"github.com/hashicorp/packer/version"
	vaultapi "github.com/hashicorp/vault/api"
)

// InitTime is the UTC time when this package was initialized. It is
// used as the timestamp for all configuration templates so that they
// match for a single build.
var InitTime time.Time

func init() {
	InitTime = time.Now().UTC()
}

// Funcs are the interpolation funcs that are available within interpolations.
var FuncGens = map[string]FuncGenerator{
	"build_name":     funcGenBuildName,
	"build_type":     funcGenBuildType,
	"env":            funcGenEnv,
	"isotime":        funcGenIsotime,
	"pwd":            funcGenPwd,
	"split":          funcGenSplitter,
	"template_dir":   funcGenTemplateDir,
	"timestamp":      funcGenTimestamp,
	"uuid":           funcGenUuid,
	"user":           funcGenUser,
	"packer_version": funcGenPackerVersion,
	"consul_key":     funcGenConsul,
	"vault":          funcGenVault,

	"upper": funcGenPrimitive(strings.ToUpper),
	"lower": funcGenPrimitive(strings.ToLower),
}

// FuncGenerator is a function that given a context generates a template
// function for the template.
type FuncGenerator func(*Context) interface{}

// Funcs returns the functions that can be used for interpolation given
// a context.
func Funcs(ctx *Context) template.FuncMap {
	result := make(map[string]interface{})
	for k, v := range FuncGens {
		result[k] = v(ctx)
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

func funcGenPrimitive(value interface{}) FuncGenerator {
	return func(ctx *Context) interface{} {
		return value
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

		return ctx.UserVariables[k], nil
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

		data := secret.Data["data"]
		if data == nil {
			return "", errors.New(fmt.Sprintf("Vault data was empty at the "+
				"given path. Warnings: %s", strings.Join(secret.Warnings, "; ")))
		}

		value := secret.Data["data"].(map[string]interface{})[key].(string)
		return value, nil
	}
}
