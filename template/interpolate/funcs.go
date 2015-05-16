package interpolate

import (
	"errors"
	"os"
	"text/template"
)

// Funcs are the interpolation funcs that are available within interpolations.
var FuncGens = map[string]FuncGenerator{
	"env":  funcGenEnv,
	"user": funcGenUser,
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

	return template.FuncMap(result)
}

func funcGenEnv(ctx *Context) interface{} {
	return func(k string) (string, error) {
		if ctx.DisableEnv {
			// The error message doesn't have to be that detailed since
			// semantic checks should catch this.
			return "", errors.New("env vars are not allowed here")
		}

		return os.Getenv(k), nil
	}
}

func funcGenUser(ctx *Context) interface{} {
	return func() string {
		return ""
	}
}
