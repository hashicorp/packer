//go:generate mapstructure-to-hcl2 -type Provisioner

package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

// Template represents the parsed template that is used to configure
// Packer builds.
type Template struct {
	// Path is the path to the template. This will be blank if Parse is
	// used, but will be automatically populated by ParseFile.
	Path string

	Description string
	MinVersion  string

	Comments           map[string]string
	Variables          map[string]*Variable
	SensitiveVariables []*Variable
	Builders           map[string]*Builder
	Provisioners       []*Provisioner
	CleanupProvisioner *Provisioner
	PostProcessors     [][]*PostProcessor

	// RawContents is just the raw data for this template
	RawContents []byte
}

// Raw converts a Template struct back into the raw Packer template structure
func (t *Template) Raw() (*rawTemplate, error) {
	var out rawTemplate

	out.MinVersion = t.MinVersion
	out.Description = t.Description

	for k, v := range t.Comments {
		out.Comments = append(out.Comments, map[string]string{k: v})
	}

	for _, b := range t.Builders {
		out.Builders = append(out.Builders, b)
	}

	for _, p := range t.Provisioners {
		out.Provisioners = append(out.Provisioners, p)
	}

	for _, pp := range t.PostProcessors {
		out.PostProcessors = append(out.PostProcessors, pp)
	}

	for _, v := range t.SensitiveVariables {
		out.SensitiveVariables = append(out.SensitiveVariables, v.Key)
	}

	for k, v := range t.Variables {
		if out.Variables == nil {
			out.Variables = make(map[string]interface{})
		}

		out.Variables[k] = v
	}

	return &out, nil
}

// Builder represents a builder configured in the template
type Builder struct {
	Name   string                 `json:"name,omitempty"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// MarshalJSON conducts the necessary flattening of the Builder struct
// to provide valid Packer template JSON
func (b *Builder) MarshalJSON() ([]byte, error) {
	// Avoid recursion
	type Builder_ Builder
	out, _ := json.Marshal(Builder_(*b))

	var m map[string]json.RawMessage
	_ = json.Unmarshal(out, &m)

	// Flatten Config
	delete(m, "config")
	for k, v := range b.Config {
		out, _ = json.Marshal(v)
		m[k] = out
	}

	return json.Marshal(m)
}

// PostProcessor represents a post-processor within the template.
type PostProcessor struct {
	OnlyExcept `mapstructure:",squash" json:",omitempty"`

	Name              string                 `json:"name,omitempty"`
	Type              string                 `json:"type"`
	KeepInputArtifact *bool                  `mapstructure:"keep_input_artifact" json:"keep_input_artifact,omitempty"`
	Config            map[string]interface{} `json:"config,omitempty"`
}

// MarshalJSON conducts the necessary flattening of the PostProcessor struct
// to provide valid Packer template JSON
func (p *PostProcessor) MarshalJSON() ([]byte, error) {
	// Early exit for simple definitions
	if len(p.Config) == 0 && len(p.OnlyExcept.Only) == 0 && len(p.OnlyExcept.Except) == 0 && p.KeepInputArtifact == nil {
		return json.Marshal(p.Type)
	}

	// Avoid recursion
	type PostProcessor_ PostProcessor
	out, _ := json.Marshal(PostProcessor_(*p))

	var m map[string]json.RawMessage
	_ = json.Unmarshal(out, &m)

	// Flatten Config
	delete(m, "config")
	for k, v := range p.Config {
		out, _ = json.Marshal(v)
		m[k] = out
	}

	return json.Marshal(m)
}

// Provisioner represents a provisioner within the template.
type Provisioner struct {
	OnlyExcept `mapstructure:",squash" json:",omitempty"`

	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Override    map[string]interface{} `json:"override,omitempty"`
	PauseBefore time.Duration          `mapstructure:"pause_before" json:"pause_before,omitempty"`
	MaxRetries  string                 `mapstructure:"max_retries" json:"max_retries,omitempty"`
	Timeout     time.Duration          `mapstructure:"timeout" json:"timeout,omitempty"`
}

// MarshalJSON conducts the necessary flattening of the Provisioner struct
// to provide valid Packer template JSON
func (p *Provisioner) MarshalJSON() ([]byte, error) {
	// Avoid recursion
	type Provisioner_ Provisioner
	out, _ := json.Marshal(Provisioner_(*p))

	var m map[string]json.RawMessage
	_ = json.Unmarshal(out, &m)

	// Flatten Config
	delete(m, "config")
	for k, v := range p.Config {
		out, _ = json.Marshal(v)
		m[k] = out
	}

	return json.Marshal(m)
}

// Push represents the configuration for pushing the template to Atlas.
type Push struct {
	Name    string
	Address string
	BaseDir string `mapstructure:"base_dir"`
	Include []string
	Exclude []string
	Token   string
	VCS     bool
}

// Variable represents a variable within the template
type Variable struct {
	Key      string
	Default  string
	Required bool
}

func (v *Variable) MarshalJSON() ([]byte, error) {
	if v.Required {
		// We use a nil pointer to coax Go into marshalling it as a JSON null
		var ret *string
		return json.Marshal(ret)
	}

	return json.Marshal(v.Default)
}

// OnlyExcept is a struct that is meant to be embedded that contains the
// logic required for "only" and "except" meta-parameters.
type OnlyExcept struct {
	Only   []string `json:"only,omitempty"`
	Except []string `json:"except,omitempty"`
}

//-------------------------------------------------------------------
// Functions
//-------------------------------------------------------------------

// Validate does some basic validation of the template on top of the
// validation that occurs while parsing. If possible, we try to defer
// validation to here. The validation errors that occur during parsing
// are the minimal necessary to make sure parsing builds a reasonable
// Template structure.
func (t *Template) Validate() error {
	var err error

	// At least one builder must be defined
	if len(t.Builders) == 0 {
		err = multierror.Append(err, errors.New(
			"at least one builder must be defined"))
	}

	// Verify that the provisioner overrides target builders that exist
	for i, p := range t.Provisioners {
		// Validate only/except
		if verr := p.OnlyExcept.Validate(t); verr != nil {
			for _, e := range multierror.Append(verr).Errors {
				err = multierror.Append(err, fmt.Errorf(
					"provisioner %d: %s", i+1, e))
			}
		}

		// Validate overrides
		for name := range p.Override {
			if _, ok := t.Builders[name]; !ok {
				err = multierror.Append(err, fmt.Errorf(
					"provisioner %d: override '%s' doesn't exist",
					i+1, name))
			}
		}
	}

	// Verify post-processors
	for i, chain := range t.PostProcessors {
		for j, p := range chain {
			// Validate only/except
			if verr := p.OnlyExcept.Validate(t); verr != nil {
				for _, e := range multierror.Append(verr).Errors {
					err = multierror.Append(err, fmt.Errorf(
						"post-processor %d.%d: %s", i+1, j+1, e))
				}
			}
		}
	}

	return err
}

// Skip says whether or not to skip the build with the given name.
func (o *OnlyExcept) Skip(n string) bool {
	if len(o.Only) > 0 {
		for _, v := range o.Only {
			if v == n {
				return false
			}
		}

		return true
	}

	if len(o.Except) > 0 {
		for _, v := range o.Except {
			if v == n {
				return true
			}
		}

		return false
	}

	return false
}

// Validate validates that the OnlyExcept settings are correct for a thing.
func (o *OnlyExcept) Validate(t *Template) error {
	if len(o.Only) > 0 && len(o.Except) > 0 {
		return errors.New("only one of 'only' or 'except' may be specified")
	}

	var err error
	for _, n := range o.Only {
		if _, ok := t.Builders[n]; !ok {
			err = multierror.Append(err, fmt.Errorf(
				"'only' specified builder '%s' not found", n))
		}
	}
	for _, n := range o.Except {
		if _, ok := t.Builders[n]; !ok {
			err = multierror.Append(err, fmt.Errorf(
				"'except' specified builder '%s' not found", n))
		}
	}

	return err
}

//-------------------------------------------------------------------
// GoStringer
//-------------------------------------------------------------------

func (b *Builder) GoString() string {
	return fmt.Sprintf("*%#v", *b)
}

func (p *Provisioner) GoString() string {
	return fmt.Sprintf("*%#v", *p)
}

func (p *PostProcessor) GoString() string {
	return fmt.Sprintf("*%#v", *p)
}

func (v *Variable) GoString() string {
	return fmt.Sprintf("*%#v", *v)
}
