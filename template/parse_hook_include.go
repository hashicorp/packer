package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"os"
	"path/filepath"
	"reflect"
)

const includeTag = "_include"

type includeResolverConfig struct {
	// Base path for includes specified with relative path.
	basePath string
}

type includeResolver struct {
	config *includeResolverConfig

	result map[string]interface{}
}

func includeHookFunc(config *includeResolverConfig) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Map {
			return data, nil
		}

		dataMapView, ok := data.(map[string]interface{})
		if !ok {
			return data, nil
		}

		return data, resolveIncludeInPlace(config, dataMapView)
	}
}

func resolveIncludeInPlace(config *includeResolverConfig, data map[string]interface{}) error {
	resolver := includeResolver{config: config, result: data}

	var errs error
	// Loop to resolve nested includes.
	// TODO Add discovering include cycles.
	for include, ok := data[includeTag]; ok; {
		delete(data, includeTag)

		err := resolver.resolve(include)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		include, ok = data[includeTag]
	}
	return errs
}

func (resolver *includeResolver) resolve(include interface{}) error {
	switch path := include.(type) {
	case string:
		return resolver.includeFromFile(path)
	case []interface{}:
		return resolver.includeFromFiles(path)
	default:
		return errors.New(fmt.Sprintf("Trying to include '%v' which is neither a string nor an array", path))
	}
}

func (resolver *includeResolver) includeFromFiles(paths []interface{}) error {
	var errs error
	for _, path := range paths {
		pathStr, ok := path.(string)
		if !ok {
			errs = multierror.Append(errs, fmt.Errorf("Trying to include '%v' which is not a string", path))
			continue
		}

		if err := resolver.includeFromFile(pathStr); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (resolver *includeResolver) includeFromFile(path string) error {
	file, err := os.Open(resolver.resolvePath(path))
	if err != nil {
		return err
	}
	defer file.Close()

	var content interface{}
	if err := json.NewDecoder(file).Decode(&content); err != nil {
		return err
	}

	contentMapView, ok := content.(map[string]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Included content is not a JSON object"))
	}

	resolver.include(contentMapView)

	return nil
}

func (resolver *includeResolver) resolvePath(path string) string {
	if resolver.config.basePath != "" && !filepath.IsAbs(path) {
		return filepath.Join(resolver.config.basePath, path)
	}
	return path
}

func (resolver *includeResolver) include(src map[string]interface{}) {
	for srcKey, srcValue := range src {
		if destValue, ok := resolver.result[srcKey]; !ok {
			// Do not overwrite value if it is already present (mitigates the risk of accidentally overwriting
			// values defined in base template by values from included templates). As a side effect, it makes
			// ordering of includes important.
			resolver.result[srcKey] = srcValue
		} else {
			if srcKey == includeTag {
				// Gather include paths from all included templates.
				resolver.result[srcKey] = appendAny(toSlice(&destValue), srcValue)
			}
		}
	}
}

func toSlice(value *interface{}) []interface{} {
	result, ok := (*value).([]interface{})
	if !ok {
		result = []interface{}{*value}
	}
	return result
}

func appendAny(dest []interface{}, value interface{}) []interface{} {
	switch value := value.(type) {
	case []interface{}:
		return append(dest, value...)
	default:
		return append(dest, value)
	}
}
