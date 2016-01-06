package config

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/template/interpolate"
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
}

// Decode decodes the configuration into the target and optionally
// automatically interpolates all the configuration as it goes.
func Decode(target interface{}, config *DecodeOpts, raws ...interface{}) error {
	if config == nil {
		config = &DecodeOpts{Interpolate: true}
	}

	// Interpolate first
	if config.Interpolate {
		// Detect user variables from the raws and merge them into our context
		ctx, err := DetectContext(raws...)
		if err != nil {
			return err
		}
		if config.InterpolateContext == nil {
			config.InterpolateContext = ctx
		} else {
			config.InterpolateContext.BuildName = ctx.BuildName
			config.InterpolateContext.BuildType = ctx.BuildType
			config.InterpolateContext.TemplatePath = ctx.TemplatePath
			config.InterpolateContext.UserVariables = ctx.UserVariables
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

	// Build our decoder
	var md mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		Metadata:         &md,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			uint8ToStringHook,
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return err
	}
	for _, raw := range raws {
		if err := decoder.Decode(raw); err != nil {
			return err
		}
	}

	// Set the metadata if it is set
	if config.Metadata != nil {
		*config.Metadata = md
	}

	// If we have unused keys, it is an error
	if len(md.Unused) > 0 {
		var err error
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			if unused != "type" && !strings.HasPrefix(unused, "packer_") {
				err = multierror.Append(err, fmt.Errorf(
					"unknown configuration key: %q", unused))
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// DetectContext builds a base interpolate.Context, automatically
// detecting things like user variables from the raw configuration params.
func DetectContext(raws ...interface{}) (*interpolate.Context, error) {
	var s struct {
		BuildName    string            `mapstructure:"packer_build_name"`
		BuildType    string            `mapstructure:"packer_builder_type"`
		TemplatePath string            `mapstructure:"packer_template_path"`
		Vars         map[string]string `mapstructure:"packer_user_variables"`
	}

	for _, r := range raws {
		if err := mapstructure.Decode(r, &s); err != nil {
			return nil, err
		}
	}

	return &interpolate.Context{
		BuildName:     s.BuildName,
		BuildType:     s.BuildType,
		TemplatePath:  s.TemplatePath,
		UserVariables: s.Vars,
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
