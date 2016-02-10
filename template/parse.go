package template

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
)

// rawTemplate is the direct JSON document format of the template file.
// This is what is decoded directly from the file, and then it is turned
// into a Template object thereafter.
type rawTemplate struct {
	MinVersion  string `mapstructure:"min_packer_version"`
	Description string

	Builders       []map[string]interface{}
	Push           map[string]interface{}
	PostProcessors []interface{} `mapstructure:"post-processors"`
	Provisioners   []map[string]interface{}
	Variables      map[string]interface{}

	RawContents []byte
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

	// Gather the variables
	if len(r.Variables) > 0 {
		result.Variables = make(map[string]*Variable, len(r.Variables))
	}
	for k, rawV := range r.Variables {
		var v Variable

		// Variable is required if the value is exactly nil
		v.Required = rawV == nil

		// Weak decode the default if we have one
		if err := r.decoder(&v.Default, nil).Decode(rawV); err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"variable %s: %s", k, err))
			continue
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
		b.Config = rawB
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

			// Set the configuration
			delete(c, "except")
			delete(c, "only")
			delete(c, "keep_input_artifact")
			delete(c, "type")
			if len(c) > 0 {
				pp.Config = c
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
		var p Provisioner
		if err := r.decoder(&p, nil).Decode(v); err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"provisioner %d: %s", i+1, err))
			continue
		}

		// Type is required before any richer validation
		if p.Type == "" {
			errs = multierror.Append(errs, fmt.Errorf(
				"provisioner %d: missing 'type'", i+1))
			continue
		}

		// Copy the configuration
		delete(v, "except")
		delete(v, "only")
		delete(v, "override")
		delete(v, "pause_before")
		delete(v, "type")
		if len(v) > 0 {
			p.Config = v
		}

		// TODO: stuff
		result.Provisioners = append(result.Provisioners, &p)
	}

	// Push
	if len(r.Push) > 0 {
		var p Push
		if err := r.decoder(&p, nil).Decode(r.Push); err != nil {
			errs = multierror.Append(errs, fmt.Errorf(
				"push: %s", err))
		}

		result.Push = p
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
	// Create a buffer to copy what we read
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	// First, decode the object into an interface{}. We do this instead of
	// the rawTemplate directly because we'd rather use mapstructure to
	// decode since it has richer errors.
	var raw interface{}
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
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
		for _, unused := range md.Unused {
			// Ignore keys starting with '_' as comments
			if unused[0] == '_' {
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

// ParseFile is the same as Parse but is a helper to automatically open
// a file for parsing.
func ParseFile(path string) (*Template, error) {
	var f *os.File
	var err error
	if path == "-" {
		// Create a temp file for stdin in case of errors
		f, err = ioutil.TempFile(os.TempDir(), "packer")
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		defer f.Close()
		io.Copy(f, os.Stdin)
		f.Seek(0, os.SEEK_SET)
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
		f.Seek(0, os.SEEK_SET)
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
