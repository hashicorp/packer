package config

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
	"github.com/ryanuber/go-glob"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// DecodeOpts are the options for decoding configuration.
type DecodeOpts struct {
	// Metadata, if non-nil, will be set to the metadata post-decode
	Metadata *mapstructure.Metadata

	// Interpolate, if true, will automatically interpolate the
	// configuration with the given InterpolateContext. User variables
	// will be automatically detected and added in-place to the given
	// context.
	Interpolate        bool
	InterpolateContext *interpolate.Context
	InterpolateFilter  *interpolate.RenderFilter

	// PluginType is the BuilderID, etc of the plugin -- it is used to
	// determine whether to tell the user to "fix" their template if an
	// unknown option is a deprecated one for this plugin type.
	PluginType string

	DecodeHooks []mapstructure.DecodeHookFunc
}

var DefaultDecodeHookFuncs = []mapstructure.DecodeHookFunc{
	uint8ToStringHook,
	stringToTrilean,
	mapstructure.StringToSliceHookFunc(","),
	mapstructure.StringToTimeDurationHookFunc(),
}

// Decode decodes the configuration into the target and optionally
// automatically interpolates all the configuration as it goes.
func Decode(target interface{}, config *DecodeOpts, raws ...interface{}) error {
	// loop over raws once to get cty values from hcl, if that's a thing.
	for i, raw := range raws {
		// check for cty values and transform them to json then to a
		// map[string]interface{} so that mapstructure can do its thing.
		cval, ok := raw.(cty.Value)
		if !ok {
			continue
		}
		type flatConfigurer interface {
			FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec }
		}
		ctarget := target.(flatConfigurer)
		flatCfg := ctarget.FlatMapstructure()
		err := gocty.FromCtyValue(cval, flatCfg)
		if err != nil {
			switch err := err.(type) {
			case cty.PathError:
				return fmt.Errorf("%v: %v", err, err.Path)
			}
			return err
		}
		b, err := ctyjson.SimpleJSONValue{Value: cval}.MarshalJSON()
		if err != nil {
			return err
		}
		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		raws[i] = raw
		{
			// reset target to zero.
			// In HCL2, we need to prepare provisioners/post-processors after a
			// builder has started in order to have build values correctly
			// extrapolated. Packer plugins have never been prepared twice in
			// the past and some of them set fields during their Validation
			// steps; which end up in an invalid provisioner/post-processor,
			// like in [GH-9596]. This ensures Packer plugin will be reset
			// right before we Prepare them.
			p := reflect.ValueOf(target).Elem()
			p.Set(reflect.Zero(p.Type()))
		}
	}

	// Now perform the normal decode.

	if config == nil {
		config = &DecodeOpts{Interpolate: true}
	}

	// Detect user variables from the raws and merge them into our context
	ctxData, raws := DetectContextData(raws...)

	// Interpolate first
	if config.Interpolate {
		ctx, err := DetectContext(raws...)
		if err != nil {
			return err
		}
		if config.InterpolateContext == nil {
			config.InterpolateContext = ctx
		} else {
			config.InterpolateContext.BuildName = ctx.BuildName
			config.InterpolateContext.BuildType = ctx.BuildType
			config.InterpolateContext.CorePackerVersionString = ctx.CorePackerVersionString
			config.InterpolateContext.TemplatePath = ctx.TemplatePath
			config.InterpolateContext.UserVariables = ctx.UserVariables
			if config.InterpolateContext.Data == nil {
				config.InterpolateContext.Data = ctxData
			}
		}
		ctx = config.InterpolateContext

		// Render everything
		for i, raw := range raws {
			m, err := interpolate.RenderMap(raw, ctx, config.InterpolateFilter)
			if err != nil {
				return err
			}

			raws[i] = m
		}
	}

	decodeHookFuncs := DefaultDecodeHookFuncs
	if len(config.DecodeHooks) != 0 {
		decodeHookFuncs = config.DecodeHooks
	}

	// Build our decoder
	var md mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		Metadata:         &md,
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(decodeHookFuncs...),
	})
	if err != nil {
		return err
	}

	// In practice, raws is two interfaces: one containing all the packer config
	// vars, and one containing the raw json configuration for a single
	// plugin.
	for _, raw := range raws {
		if err := decoder.Decode(raw); err != nil {
			return err
		}
	}

	// If we have unused keys, it is an error
	if len(md.Unused) > 0 {
		var err error
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			if unused == "type" || strings.HasPrefix(unused, "packer_") {
				continue
			}

			// Check for whether the key is handled in a packer fix
			// call.
			fixable := false

			// check whether the deprecation option can be fixed using packer fix.
			if config.PluginType != "" {
				for k, deprecatedOptions := range DeprecatedOptions {
					// the deprecated options keys are globbable, for
					// example "amazon*" for all amazon builders, or * for
					// all builders
					if glob.Glob(k, config.PluginType) {
						for _, deprecatedOption := range deprecatedOptions {
							if unused == deprecatedOption {
								fixable = true
								break
							}
						}
					}
					if fixable == true {
						break
					}
				}
			}

			unusedErr := fmt.Errorf("unknown configuration key: '%q'",
				unused)

			if fixable {
				unusedErr = fmt.Errorf("Deprecated configuration key: '%s'."+
					" Please call `packer fix` against your template to "+
					"update your template to be compatible with the current "+
					"version of Packer. Visit "+
					"https://www.packer.io/docs/commands/fix/ for more detail.",
					unused)
			}

			err = multierror.Append(err, unusedErr)
		}
		if err != nil {
			return err
		}
	}

	// Set the metadata if it is set
	if config.Metadata != nil {
		*config.Metadata = md
	}

	return nil
}

func DetectContextData(raws ...interface{}) (map[interface{}]interface{}, []interface{}) {
	// In provisioners, the last value pulled from raws is the placeholder data
	// for build-specific variables. Pull these out to add to interpolation
	// context.
	if len(raws) == 0 {
		return nil, raws
	}

	// Internally, our tests may cause this to be read as a map[string]string
	placeholderData := raws[len(raws)-1]
	if pd, ok := placeholderData.(map[string]string); ok {
		if uuid, ok := pd["PackerRunUUID"]; ok {
			if strings.Contains(uuid, "Build_PackerRunUUID.") {
				cast := make(map[interface{}]interface{})
				for k, v := range pd {
					cast[k] = v
				}
				raws = raws[:len(raws)-1]
				return cast, raws
			}
		}
	}

	// but with normal interface conversion across the rpc, it'll look like a
	// map[interface]interface, not a map[string]string
	if pd, ok := placeholderData.(map[interface{}]interface{}); ok {
		if uuid, ok := pd["PackerRunUUID"]; ok {
			if strings.Contains(uuid.(string), "Build_PackerRunUUID.") {
				raws = raws[:len(raws)-1]
				return pd, raws
			}
		}
	}

	return nil, raws
}

// DetectContext builds a base interpolate.Context, automatically
// detecting things like user variables from the raw configuration params.
func DetectContext(raws ...interface{}) (*interpolate.Context, error) {
	var s struct {
		BuildName               string            `mapstructure:"packer_build_name"`
		BuildType               string            `mapstructure:"packer_builder_type"`
		CorePackerVersionString string            `mapstructure:"packer_core_version"`
		TemplatePath            string            `mapstructure:"packer_template_path"`
		Vars                    map[string]string `mapstructure:"packer_user_variables"`
		SensitiveVars           []string          `mapstructure:"packer_sensitive_variables"`
	}

	for _, r := range raws {
		if err := mapstructure.Decode(r, &s); err != nil {
			log.Printf("Error detecting context: %s", err)
			return nil, err
		}
	}

	return &interpolate.Context{
		BuildName:               s.BuildName,
		BuildType:               s.BuildType,
		CorePackerVersionString: s.CorePackerVersionString,
		TemplatePath:            s.TemplatePath,
		UserVariables:           s.Vars,
		SensitiveVariables:      s.SensitiveVars,
	}, nil
}

func uint8ToStringHook(f reflect.Kind, t reflect.Kind, v interface{}) (interface{}, error) {
	// We need to convert []uint8 to string. We have to do this
	// because internally Packer uses MsgPack for RPC and the MsgPack
	// codec turns strings into []uint8
	if f == reflect.Slice && t == reflect.String {
		dataVal := reflect.ValueOf(v)
		dataType := dataVal.Type()
		elemKind := dataType.Elem().Kind()
		if elemKind == reflect.Uint8 {
			v = string(dataVal.Interface().([]uint8))
		}
	}

	return v, nil
}

func stringToTrilean(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {
	// We have a custom data type, config, which we read from a string and
	// then cast to a *bool. Why? So that we can appropriately read "unset"
	// *bool values in order to intelligently default, even when the values are
	// being set by a template variable.

	testTril, _ := TrileanFromString("")
	if t == reflect.TypeOf(testTril) {
		// From value is string
		if f == reflect.TypeOf("") {
			tril, err := TrileanFromString(v.(string))
			if err != nil {
				return v, fmt.Errorf("Error parsing bool from given var: %s", err)
			}
			return tril, nil
		} else {
			// From value is boolean
			if f == reflect.TypeOf(true) {
				tril := TrileanFromBool(v.(bool))
				return tril, nil
			}
		}

	}
	return v, nil
}
