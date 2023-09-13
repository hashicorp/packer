package json

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func isDoneInterpolating(v string) (bool, error) {
	// Check for whether the var contains any more references to `user`, wrapped
	// in interpolation syntax.
	filter := `{{\s*user\s*\x60.*\x60\s*}}`
	matched, err := regexp.MatchString(filter, v)
	if err != nil {
		return false, fmt.Errorf("Can't tell if interpolation is done: %s", err)
	}
	if matched {
		// not done interpolating; there's still a call to "user" in a template
		// engine
		return false, nil
	}
	// No more calls to "user" as a template engine, so we're done.
	return true, nil
}

func (s *JSONSequentialScheduler) renderVarsRecursively() (*interpolate.Context, error) {
	ctx := s.config.Context()
	ctx.EnableEnv = true
	ctx.UserVariables = make(map[string]string)
	shouldRetry := true
	changed := false
	failedInterpolation := ""

	// Why this giant loop?  User variables can be recursively defined. For
	// example:
	// "variables": {
	//    	"foo":  "bar",
	//	 	"baz":  "{{user `foo`}}baz",
	// 		"bang": "bang{{user `baz`}}"
	// },
	// In this situation, we cannot guarantee that we've added "foo" to
	// UserVariables before we try to interpolate "baz" the first time. We need
	// to have the option to loop back over in order to add the properly
	// interpolated "baz" to the UserVariables map.
	// Likewise, we'd need to loop up to two times to properly add "bang",
	// since that depends on "baz" being set, which depends on "foo" being set.

	// We break out of the while loop either if all our variables have been
	// interpolated or if after 100 loops we still haven't succeeded in
	// interpolating them.  Please don't actually nest your variables in 100
	// layers of other variables. Please.

	// c.Template.Variables is populated by variables defined within the Template
	// itself
	// c.variables is populated by variables read in from the command line and
	// var-files.
	// We need to read the keys from both, then loop over all of them to figure
	// out the appropriate interpolations.

	repeatMap := make(map[string]string)
	allKeys := make([]string, 0)

	// load in template variables
	for k, v := range s.config.Template.Variables {
		repeatMap[k] = v.Default
		allKeys = append(allKeys, k)
	}

	// overwrite template variables with command-line-read variables
	for k, v := range s.config.Variables {
		repeatMap[k] = v
		allKeys = append(allKeys, k)
	}

	// sort map to force the following loop to be deterministic.
	sort.Strings(allKeys)
	type keyValue struct {
		Key   string
		Value string
	}
	sortedMap := make([]keyValue, len(repeatMap))
	for _, k := range allKeys {
		sortedMap = append(sortedMap, keyValue{k, repeatMap[k]})
	}

	// Regex to exclude any build function variable or template variable
	// from interpolating earlier
	// E.g.: {{ .HTTPIP }}  won't interpolate now
	renderFilter := "{{(\\s|)\\.(.*?)(\\s|)}}"

	for i := 0; i < 100; i++ {
		shouldRetry = false
		changed = false
		deleteKeys := []string{}
		// First, loop over the variables in the template
		for _, kv := range sortedMap {
			// Interpolate the default
			renderedV, err := interpolate.RenderRegex(kv.Value, ctx, renderFilter)
			switch err.(type) {
			case nil:
				// We only get here if interpolation has succeeded, so something is
				// different in this loop than in the last one.
				changed = true
				s.config.Variables[kv.Key] = renderedV
				ctx.UserVariables = s.config.Variables
				// Remove fully-interpolated variables from the map, and flag
				// variables that still need interpolating for a repeat.
				done, err := isDoneInterpolating(kv.Value)
				if err != nil {
					return ctx, err
				}
				if done {
					deleteKeys = append(deleteKeys, kv.Key)
				} else {
					shouldRetry = true
				}
			case template.ExecError:
				castError := err.(template.ExecError)
				if strings.Contains(castError.Error(), interpolate.ErrVariableNotSetString) {
					shouldRetry = true
					failedInterpolation = fmt.Sprintf(`"%s": "%s"; error: %s`, kv.Key, kv.Value, err)
				} else {
					return ctx, err
				}
			default:
				return ctx, fmt.Errorf(
					// unexpected interpolation error: abort the run
					"error interpolating default value for '%s': %s",
					kv.Key, err)
			}
		}
		if !shouldRetry {
			break
		}

		// Clear completed vars from sortedMap before next loop. Do this one
		// key at a time because the indices are gonna change ever time you
		// delete from the map.
		for _, k := range deleteKeys {
			for ind, kv := range sortedMap {
				if kv.Key == k {
					sortedMap = append(sortedMap[:ind], sortedMap[ind+1:]...)
					break
				}
			}
		}
		deleteKeys = []string{}
	}

	if !changed && shouldRetry {
		return ctx, fmt.Errorf("Failed to interpolate %s: Please make sure that "+
			"the variable you're referencing has been defined; Packer treats "+
			"all variables used to interpolate other user variables as "+
			"required.", failedInterpolation)
	}

	return ctx, nil
}

func (s *JSONSequentialScheduler) getEvaluatedVariables() ([]string, error) {
	if s.config.Variables == nil {
		s.config.Variables = make(map[string]string)
	}
	// Go through the variables and interpolate the environment and
	// user variables
	ctx, err := s.renderVarsRecursively()
	if err != nil {
		return nil, err
	}

	secrets := []string{}

	for _, v := range s.config.Template.SensitiveVariables {
		secret := ctx.UserVariables[v.Key]
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func (s *JSONSequentialScheduler) EvaluateVariables() hcl.Diagnostics {
	secrets, err := s.getEvaluatedVariables()

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to evaluate variables",
				Detail:   err.Error(),
			},
		}
	}
	for _, secret := range secrets {
		packersdk.LogSecretFilter.Set(secret)
	}

	return nil
}
