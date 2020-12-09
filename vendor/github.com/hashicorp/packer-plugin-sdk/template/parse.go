package template

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/mitchellh/mapstructure"
)

// rawTemplate is the direct JSON document format of the template file.
// This is what is decoded directly from the file, and then it is turned
// into a Template object thereafter.
type rawTemplate struct {
	MinVersion  string `mapstructure:"min_packer_version" json:"min_packer_version,omitempty"`
	Description string `json:"description,omitempty"`

	Builders           []interface{}          `mapstructure:"builders" json:"builders,omitempty"`
	Comments           []map[string]string    `json:"comments,omitempty"`
	Push               map[string]interface{} `json:"push,omitempty"`
	PostProcessors     []interface{}          `mapstructure:"post-processors" json:"post-processors,omitempty"`
	Provisioners       []interface{}          `json:"provisioners,omitempty"`
	CleanupProvisioner interface{}            `mapstructure:"error-cleanup-provisioner" json:"error-cleanup-provisioner,omitempty"`
	Variables          map[string]interface{} `json:"variables,omitempty"`
	SensitiveVariables []string               `mapstructure:"sensitive-variables" json:"sensitive-variables,omitempty"`

	RawContents []byte `json:"-"`
}

// MarshalJSON conducts the necessary flattening of the rawTemplate struct
// to provide valid Packer template JSON
func (r *rawTemplate) MarshalJSON() ([]byte, error) {
	// Avoid recursion
	type rawTemplate_ rawTemplate
	out, _ := json.Marshal(rawTemplate_(*r))

	var m map[string]json.RawMessage
	_ = json.Unmarshal(out, &m)

	// Flatten Comments
	delete(m, "comments")
	for _, comment := range r.Comments {
		for k, v := range comment {
			out, _ = json.Marshal(v)
			m[k] = out
		}
	}

	return json.Marshal(m)
}

func (r *rawTemplate) decodeProvisioner(raw interface{}) (Provisioner, error) {
	var p Provisioner
	if err := r.weakDecoder(&p, nil).Decode(raw); err != nil {
		return p, fmt.Errorf("Error decoding provisioner: %s", err)

	}

	// Type is required before any richer validation
	if p.Type == "" {
		return p, fmt.Errorf("Provisioner missing 'type'")
	}

	// Set the raw configuration and delete any special keys
	p.Config = raw.(map[string]interface{})

	delete(p.Config, "except")
	delete(p.Config, "only")
	delete(p.Config, "override")
	delete(p.Config, "pause_before")
	delete(p.Config, "max_retries")
	delete(p.Config, "type")
	delete(p.Config, "timeout")

	if len(p.Config) == 0 {
		p.Config = nil
	}
	return p, nil
}

// Template returns the actual Template object built from this raw
// structure.
func (r *rawTemplate) Template() (*Template, error) {
	var result Template
	var errs error

	// Copy some literals
	result.Description = r.Description
	result.MinVersion = r.MinVersion
	result.RawContents = r.RawContents

	// Gather the comments
	if len(r.Comments) > 0 {
		result.Comments = make(map[string]string, len(r.Comments))

		for _, c := range r.Comments {
			for k, v := range c {
				result.Comments[k] = v
			}
		}
	}

	// Gather the variables
	if len(r.Variables) > 0 {
		result.Variables = make(map[string]*Variable, len(r.Variables))
	}

	for k, rawV := range r.Variables {
		var v Variable
		v.Key = k

		// Variable is required if the value is exactly nil
		v.Required = rawV == nil

		// Weak decode the default if we have one
		if err := r.decoder(&v.Default, nil).Decode(rawV); err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"variable %s: %s", k, err))
			continue
		}

		for _, sVar := range r.SensitiveVariables {
			if sVar == k {
				result.SensitiveVariables = append(result.SensitiveVariables, &v)
			}
		}

		result.Variables[k] = &v
	}

	// Let's start by gathering all the builders
	if len(r.Builders) > 0 {
		result.Builders = make(map[string]*Builder, len(r.Builders))
	}
	for i, rawB := range r.Builders {
		var b Builder
		if err := mapstructure.WeakDecode(rawB, &b); err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"builder %d: %s", i+1, err))
			continue
		}

		// Set the raw configuration and delete any special keys
		b.Config = rawB.(map[string]interface{})

		delete(b.Config, "name")
		delete(b.Config, "type")

		if len(b.Config) == 0 {
			b.Config = nil
		}

		// If there is no type set, it is an error
		if b.Type == "" {
			errs = multierror.Append(errs, fmt.Errorf(
				"builder %d: missing 'type'", i+1))
			continue
		}

		// The name defaults to the type if it isn't set
		if b.Name == "" {
			b.Name = b.Type
		}

		// If this builder already exists, it is an error
		if _, ok := result.Builders[b.Name]; ok {
			errs = multierror.Append(errs, fmt.Errorf(
				"builder %d: builder with name '%s' already exists",
				i+1, b.Name))
			continue
		}

		// Append the builders
		result.Builders[b.Name] = &b
	}

	// Gather all the post-processors
	if len(r.PostProcessors) > 0 {
		result.PostProcessors = make([][]*PostProcessor, 0, len(r.PostProcessors))
	}
	for i, v := range r.PostProcessors {
		// Parse the configurations. We need to do this because post-processors
		// can take three different formats.
		configs, err := r.parsePostProcessor(i, v)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		// Parse the PostProcessors out of the configs
		pps := make([]*PostProcessor, 0, len(configs))
		for j, c := range configs {
			var pp PostProcessor
			if err := r.decoder(&pp, nil).Decode(c); err != nil {
				errs = multierror.Append(errs, fmt.Errorf(
					"post-processor %d.%d: %s", i+1, j+1, err))
				continue
			}

			// Type is required
			if pp.Type == "" {
				errs = multierror.Append(errs, fmt.Errorf(
					"post-processor %d.%d: type is required", i+1, j+1))
				continue
			}

			// Set the raw configuration and delete any special keys
			pp.Config = c

			// The name defaults to the type if it isn't set
			if pp.Name == "" {
				pp.Name = pp.Type
			}

			delete(pp.Config, "except")
			delete(pp.Config, "only")
			delete(pp.Config, "keep_input_artifact")
			delete(pp.Config, "type")
			delete(pp.Config, "name")

			if len(pp.Config) == 0 {
				pp.Config = nil
			}

			pps = append(pps, &pp)
		}

		result.PostProcessors = append(result.PostProcessors, pps)
	}

	// Gather all the provisioners
	if len(r.Provisioners) > 0 {
		result.Provisioners = make([]*Provisioner, 0, len(r.Provisioners))
	}
	for i, v := range r.Provisioners {
		p, err := r.decodeProvisioner(v)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"provisioner %d: %s", i+1, err))
			continue
		}

		result.Provisioners = append(result.Provisioners, &p)
	}

	// Gather the error-cleanup-provisioner
	if r.CleanupProvisioner != nil {
		p, err := r.decodeProvisioner(r.CleanupProvisioner)
		if err != nil {
			errs = multierror.Append(errs,
				fmt.Errorf("On Error Cleanup Provisioner error: %s", err))
		}

		result.CleanupProvisioner = &p
	}

	// If we have errors, return those with a nil result
	if errs != nil {
		return nil, errs
	}

	return &result, nil
}

func (r *rawTemplate) decoder(
	result interface{},
	md *mapstructure.Metadata) *mapstructure.Decoder {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Metadata:   md,
		Result:     result,
	})
	if err != nil {
		// This really shouldn't happen since we have firm control over
		// all the arguments and they're all unit tested. So we use a
		// panic here to note this would definitely be a bug.
		panic(err)
	}
	return d
}

func (r *rawTemplate) weakDecoder(
	result interface{},
	md *mapstructure.Metadata) *mapstructure.Decoder {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		Metadata:         md,
		Result:           result,
	})
	if err != nil {
		// This really shouldn't happen since we have firm control over
		// all the arguments and they're all unit tested. So we use a
		// panic here to note this would definitely be a bug.
		panic(err)
	}
	return d
}

func (r *rawTemplate) parsePostProcessor(
	i int, raw interface{}) ([]map[string]interface{}, error) {
	switch v := raw.(type) {
	case string:
		return []map[string]interface{}{
			{"type": v},
		}, nil
	case map[string]interface{}:
		return []map[string]interface{}{v}, nil
	case []interface{}:
		var err error
		result := make([]map[string]interface{}, len(v))
		for j, innerRaw := range v {
			switch innerV := innerRaw.(type) {
			case string:
				result[j] = map[string]interface{}{"type": innerV}
			case map[string]interface{}:
				result[j] = innerV
			case []interface{}:
				err = multierror.Append(err, fmt.Errorf(
					"post-processor %d.%d: sequence not allowed to be nested in a sequence",
					i+1, j+1))
			default:
				err = multierror.Append(err, fmt.Errorf(
					"post-processor %d.%d: unknown format",
					i+1, j+1))
			}
		}

		if err != nil {
			return nil, err
		}

		return result, nil
	default:
		return nil, fmt.Errorf("post-processor %d: bad format", i+1)
	}
}

// Parse takes the given io.Reader and parses a Template object out of it.
func Parse(r io.Reader) (*Template, error) {
	// First, decode the object into an interface{} and search for duplicate fields.
	// We do this instead of the rawTemplate directly because we'd rather use mapstructure to
	// decode since it has richer errors.
	var raw interface{}
	buf, err := jsonUnmarshal(r, &raw)
	if err != nil {
		return nil, err
	}

	// Create our decoder
	var md mapstructure.Metadata
	var rawTpl rawTemplate
	rawTpl.RawContents = buf.Bytes()
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &rawTpl,
	})
	if err != nil {
		return nil, err
	}

	// Do the actual decode into our structure
	if err := decoder.Decode(raw); err != nil {
		return nil, err
	}

	// Build an error if there are unused root level keys
	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)

		unusedMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Failed to convert unused root level keys to map")
		}

		for _, unused := range md.Unused {
			if unused[0] == '_' {
				commentVal, ok := unusedMap[unused].(string)
				if !ok {
					return nil, fmt.Errorf("Failed to cast root level comment value in comment \"%s\" to string.", unused)
				}

				comment := map[string]string{
					unused: commentVal,
				}

				rawTpl.Comments = append(rawTpl.Comments, comment)
				continue
			}

			err = multierror.Append(err, fmt.Errorf(
				"Unknown root level key in template: '%s'", unused))
		}
	}
	if err != nil {
		return nil, err
	}

	// Return the template parsed from the raw structure
	return rawTpl.Template()
}

func jsonUnmarshal(r io.Reader, raw *interface{}) (bytes.Buffer, error) {
	// Create a buffer to copy what we read
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return buf, err
	}

	// Decode the object into an interface{}
	if err := json.Unmarshal(buf.Bytes(), raw); err != nil {
		return buf, err
	}

	// If Json is valid, check for duplicate fields to avoid silent unwanted override
	jsonDecoder := json.NewDecoder(strings.NewReader(buf.String()))
	if err := checkForDuplicateFields(jsonDecoder); err != nil {
		return buf, err
	}

	return buf, nil
}

func checkForDuplicateFields(d *json.Decoder) error {
	// Get next token from JSON
	t, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	// Do nothing if it's not a delimiter
	if !ok {
		return nil
	}

	// Check for duplicates inside of a delimiter {} or []
	switch delim {
	case '{':
		keys := make(map[string]bool)
		for d.More() {
			// Get attribute key
			t, err := d.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			// Check for duplicates
			if keys[key] {
				return fmt.Errorf("template has duplicate field: %s", key)
			}
			keys[key] = true

			// Check value to find duplicates in nested blocks
			if err := checkForDuplicateFields(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := checkForDuplicateFields(d); err != nil {
				return err
			}
		}
	}

	// consume closing delimiter } or ]
	if _, err := d.Token(); err != nil {
		return err
	}

	return nil
}

// ParseFile is the same as Parse but is a helper to automatically open
// a file for parsing.
func ParseFile(path string) (*Template, error) {
	var f *os.File
	var err error
	if path == "-" {
		// Create a temp file for stdin in case of errors
		f, err = tmp.File("parse")
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		defer f.Close()
		if _, err = io.Copy(f, os.Stdin); err != nil {
			return nil, err
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	} else {
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}
	tpl, err := Parse(f)
	if err != nil {
		syntaxErr, ok := err.(*json.SyntaxError)
		if !ok {
			return nil, err
		}
		// Rewind the file and get a better error
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		// Grab the error location, and return a string to point to offending syntax error
		line, col, highlight := highlightPosition(f, syntaxErr.Offset)
		err = fmt.Errorf("Error parsing JSON: %s\nAt line %d, column %d (offset %d):\n%s", err, line, col, syntaxErr.Offset, highlight)
		return nil, err
	}

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}

	tpl.Path = path
	return tpl, nil
}

// Takes a file and the location in bytes of a parse error
// from json.SyntaxError.Offset and returns the line, column,
// and pretty-printed context around the error with an arrow indicating the exact
// position of the syntax error.
func highlightPosition(f *os.File, pos int64) (line, col int, highlight string) {
	// Modified version of the function in Camlistore by Brad Fitzpatrick
	// https://github.com/camlistore/camlistore/blob/4b5403dd5310cf6e1ae8feb8533fd59262701ebc/vendor/go4.org/errorutil/highlight.go
	line = 1
	// New io.Reader for file
	br := bufio.NewReader(f)
	// Initialize lines
	lastLine := ""
	thisLine := new(bytes.Buffer)
	// Loop through template to find line, column
	for n := int64(0); n < pos; n++ {
		// read byte from io.Reader
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		// If end of line, save line as previous line in case next line is offender
		if b == '\n' {
			lastLine = thisLine.String()
			thisLine.Reset()
			line++
			col = 1
		} else {
			// Write current line, until line is safe, or error point is encountered
			col++
			thisLine.WriteByte(b)
		}
	}

	// Populate highlight string to place a '^' char at offending column
	if line > 1 {
		highlight += fmt.Sprintf("%5d: %s\n", line-1, lastLine)
	}

	highlight += fmt.Sprintf("%5d: %s\n", line, thisLine.String())
	highlight += fmt.Sprintf("%s^\n", strings.Repeat(" ", col+5))
	return
}
