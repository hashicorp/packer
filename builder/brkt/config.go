package brkt

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/packer/template/interpolate"
)

// MachineTypeConfig configures which machine type this will eventually be
// deployed onto.
type MachineTypeConfig struct {
	MinCpuCores int     `mapstructure:"min_cpu_cores"`
	MinRam      float64 `mapstructure:"min_ram_in_gb"`
	MachineType string  `mapstructure:"machine_type_uuid"`
}

func (c *MachineTypeConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	minimumsSpecified := c.MinCpuCores != 0 && c.MinRam != 0
	machineTypeSpecified := c.MachineType != ""

	if !minimumsSpecified && !machineTypeSpecified {
		errs = append(errs, fmt.Errorf("must specify either a machine_type_uuid or min_ram_in_gb and min_cpu_cores"))
	}

	return errs
}

type WorkloadConfig struct {
	ImageDefinition  string                 `mapstructure:"image_definition_uuid"`
	BillingGroup     string                 `mapstructure:"billing_group_uuid"`
	Zone             string                 `mapstructure:"zone_uuid"`
	SecurityGroup    string                 `mapstructure:"security_group_uuid"`
	CloudConfig      map[string]interface{} `mapstructure:"cloud_config"`
	MetavisorEnabled bool                   `mapstructure:"metavisor_enabled"`
}

// convertVal, convertSlice and convertMap are used to change the embedded
// map[interface{}]interface{} objects to map[string]interface{} to allow them
// to be marshaled into JSON later
// NOTE: the early declarations of `val err error` below are due to a Golang issue
// 		 which is documented here: https://github.com/golang/go/issues/6842
func convertVal(val interface{}) (interface{}, error) {
	switch v := val.(type) {
	case []interface{}:
		return convertSlice(v)
	case map[interface{}]interface{}:
		return convertMap(v)
	case interface{}:
		return v, nil
	}

	// this is just here because the method needs a return at the bottom,
	// the `case interface{}:` should catch all
	return nil, fmt.Errorf("the value `%+v` could not be converted", val)
}

func convertSlice(slice []interface{}) ([]interface{}, error) {
	retval := make([]interface{}, len(slice))

	for i, val := range slice {
		var err error
		retval[i], err = convertVal(val)
		if err != nil {
			return nil, err
		}
	}

	return retval, nil
}

func convertMap(data map[interface{}]interface{}) (map[string]interface{}, error) {
	retval := make(map[string]interface{}, len(data))

	for key, val := range data {
		k, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("error converting JSON key to string")
		}

		var err error
		retval[k], err = convertVal(val)
		if err != nil {
			return nil, err
		}
	}

	return retval, nil
}

// convert is used to begin converting since the root of CloudConfig is a
// map[string]interface{} and not a map[interface{}]interface{}. Just setting
// its' type to map[interface{}]interface{} in the config isn't good either
// since we access later CloudConfig[ssh_authorized_keys] (step_deploy_instance)
func convert(data map[string]interface{}) (map[string]interface{}, error) {
	for key, val := range data {
		var err error
		data[key], err = convertVal(val)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (c *WorkloadConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	// required arguments
	if c.ImageDefinition == "" {
		errs = append(errs, fmt.Errorf("image_definition_uuid required"))
	}
	if c.BillingGroup == "" {
		errs = append(errs, fmt.Errorf("billing_group_uuid required"))
	}
	if c.Zone == "" {
		errs = append(errs, fmt.Errorf("zone_uuid required"))
	}

	// optional argument
	if c.CloudConfig != nil {
		var err error
		c.CloudConfig, err = convert(c.CloudConfig)
		if err != nil {
			errs = append(errs, err)
		}
	} else {
		// init empty map
		c.CloudConfig = make(map[string]interface{}, 0)
	}

	return errs
}

// ImageConfig holds the configuration options relating to the artifact
// eventually output by this builder.
type ImageConfig struct {
	ImageName string `mapstructure:"image_name"`
}

func (c *ImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("image_name required"))
	}

	return errs
}

// AccessConfig holds configuration options relating to authenticating against
// the portal in order to make our API requests.
type AccessConfig struct {
	PortalUrl   string `mapstructure:"portal_url"`
	AccessToken string `mapstructure:"access_token"`
	MacKey      string `mapstructure:"mac_key"`
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.PortalUrl == "" {
		c.PortalUrl = brkt.DEFAULT_PORTAL_URL
	}

	match, err := regexp.Match("^https?://", []byte(c.PortalUrl))
	if err != nil {
		errs = append(errs, fmt.Errorf("error regexing portal_url: %s", err))
	}
	if !match {
		errs = append(errs, fmt.Errorf("please add http:// or https:// to your portal_url"))
	}

	if trailingSlash := strings.HasSuffix(c.PortalUrl, "/"); !trailingSlash {
		c.PortalUrl = c.PortalUrl + "/"
	}

	// trim quotes that might have snuck into the variables from the command line
	c.AccessToken = strings.Trim(c.AccessToken, "\"")
	c.MacKey = strings.Trim(c.MacKey, "\"")

	if c.AccessToken == "" {
		errs = append(errs, fmt.Errorf("access_token required"))
	}
	if c.MacKey == "" {
		errs = append(errs, fmt.Errorf("mac_key required"))
	}

	return errs
}
