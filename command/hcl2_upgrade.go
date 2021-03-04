package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	texttemplate "text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclwrite"
	hcl2shim "github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"
	"github.com/zclconf/go-cty/cty"
)

type HCL2UpgradeCommand struct {
	Meta
}

func (c *HCL2UpgradeCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *HCL2UpgradeCommand) ParseArgs(args []string) (*HCL2UpgradeArgs, int) {
	var cfg HCL2UpgradeArgs
	flags := c.Meta.FlagSet("hcl2_upgrade", FlagSetNone)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}
	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.Path = args[0]
	if cfg.OutputFile == "" {
		cfg.OutputFile = cfg.Path + ".pkr.hcl"
	}
	return &cfg, 0
}

const (
	hcl2UpgradeFileHeader = `# This file was autogenerated by the 'packer hcl2_upgrade' command. We
# recommend double checking that everything is correct before going forward. We
# also recommend treating this file as disposable. The HCL2 blocks in this
# file can be moved to other files. For example, the variable blocks could be
# moved to their own 'variables.pkr.hcl' file, etc. Those files need to be
# suffixed with '.pkr.hcl' to be visible to Packer. To use multiple files at
# once they also need to be in the same folder. 'packer inspect folder/'
# will describe to you what is in that folder.

# Avoid mixing go templating calls ( for example ` + "```{{ upper(`string`) }}```" + ` )
# and HCL2 calls (for example '${ var.string_value_example }' ). They won't be
# executed together and the outcome will be unknown.
`
	inputVarHeader = `
# All generated input variables will be of 'string' type as this is how Packer JSON
# views them; you can change their type later on. Read the variables type
# constraints documentation
# https://www.packer.io/docs/templates/hcl_templates/variables#type-constraints for more info.`
	localsVarHeader = `
# All locals variables are generated from variables that uses expressions
# that are not allowed in HCL2 variables.
# Read the documentation for locals blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/locals`
	packerBlockHeader = `
# See https://www.packer.io/docs/templates/hcl_templates/blocks/packer for more info
`

	sourcesHeader = `
# source blocks are generated from your builders; a source can be referenced in
# build blocks. A build block runs provisioner and post-processors on a
# source. Read the documentation for source blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/source`

	buildHeader = `
# a build block invokes sources and runs provisioning steps on them. The
# documentation for build blocks can be found here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/build
`

	amazonAmiDataHeader = `
# The amazon-ami data block is generated from your amazon builder source_ami_filter; a data
# from this block can be referenced in source and locals blocks.
# Read the documentation for data blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/data
# Read the documentation for the Amazon AMI Data Source here:
# https://www.packer.io/docs/datasources/amazon/ami`

	amazonSecretsManagerDataHeader = `
# The amazon-secretsmanager data block is generated from your aws_secretsmanager template function; a data
# from this block can be referenced in source and locals blocks.
# Read the documentation for data blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/data
# Read the documentation for the Amazon Secrets Manager Data Source here:
# https://www.packer.io/docs/datasources/amazon/secretsmanager`
)

var (
	amazonSecretsManagerMap = map[string]map[string]interface{}{}
	localsVariableMap       = map[string]string{}
	timestamp               = false
)

type BlockParser interface {
	Parse(*template.Template) error
	Write(*bytes.Buffer)
}

func (c *HCL2UpgradeCommand) RunContext(_ context.Context, cla *HCL2UpgradeArgs) int {
	var output io.Writer
	if err := os.MkdirAll(filepath.Dir(cla.OutputFile), 0); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to create output directory: %v", err))
		return 1
	}
	if f, err := os.Create(cla.OutputFile); err == nil {
		output = f
		defer f.Close()
	} else {
		c.Ui.Error(fmt.Sprintf("Failed to create output file: %v", err))
		return 1
	}

	if cla.WithAnnotations {
		if _, err := output.Write([]byte(hcl2UpgradeFileHeader)); err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to write to file: %v", err))
			return 1
		}
	}

	hdl, ret := c.GetConfigFromJSON(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	core := hdl.(*CoreWrapper).Core
	if err := core.Initialize(); err != nil {
		c.Ui.Error(fmt.Sprintf("Ignoring following initialization error: %v", err))
	}
	tpl := core.Template

	// Parse blocks

	packerBlock := &PackerParser{
		WithAnnotations: cla.WithAnnotations,
	}
	if err := packerBlock.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	variables := &VariableParser{
		WithAnnotations: cla.WithAnnotations,
	}
	if err := variables.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	locals := &LocalsParser{
		LocalsOut:       variables.localsOut,
		WithAnnotations: cla.WithAnnotations,
	}
	if err := locals.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	builders := []*template.Builder{}
	{
		// sort builders to avoid map's randomnes
		for _, builder := range tpl.Builders {
			builders = append(builders, builder)
		}
	}
	sort.Slice(builders, func(i, j int) bool {
		return builders[i].Type+builders[i].Name < builders[j].Type+builders[j].Name
	})

	amazonAmiDatasource := &AmazonAmiDatasourceParser{
		Builders:        builders,
		WithAnnotations: cla.WithAnnotations,
	}
	if err := amazonAmiDatasource.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	sources := &SourceParser{
		Builders:        builders,
		BuilderPlugins:  c.Meta.CoreConfig.Components.PluginConfig.Builders,
		WithAnnotations: cla.WithAnnotations,
	}
	if err := sources.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	build := &BuildParser{
		Builders:        builders,
		WithAnnotations: cla.WithAnnotations,
	}
	if err := build.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	amazonSecretsDatasource := &AmazonSecretsDatasourceParser{
		WithAnnotations: cla.WithAnnotations,
	}
	if err := amazonSecretsDatasource.Parse(tpl); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// Write file
	out := &bytes.Buffer{}
	for _, block := range []BlockParser{
		packerBlock,
		variables,
		amazonSecretsDatasource,
		amazonAmiDatasource,
		locals,
		sources,
		build,
	} {
		block.Write(out)
	}

	if _, err := output.Write(hclwrite.Format(out.Bytes())); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to write to file: %v", err))
		return 1
	}

	c.Ui.Say(fmt.Sprintf("Successfully created %s ", cla.OutputFile))
	return 0
}

type UnhandleableArgumentError struct {
	Call           string
	Correspondance string
	Docs           string
}

func (uc UnhandleableArgumentError) Error() string {
	return fmt.Sprintf(`unhandled %q call:
# there is no way to automatically upgrade the %[1]q call.
# Please manually upgrade to %s
# Visit %s for more infos.`, uc.Call, uc.Correspondance, uc.Docs)
}

// transposeTemplatingCalls executes parts of blocks as go template files and replaces
// their result with their hcl2 variant. If something goes wrong the template
// containing the go template string is returned.
func transposeTemplatingCalls(s []byte) []byte {
	fallbackReturn := func(err error) []byte {
		if strings.Contains(err.Error(), "unhandled") {
			return append([]byte(fmt.Sprintf("\n# %s\n", err)), s...)
		}

		return append([]byte(fmt.Sprintf("\n# could not parse template for following block: %q\n", err)), s...)
	}

	funcErrors := &multierror.Error{
		ErrorFormat: func(es []error) string {
			if len(es) == 1 {
				return fmt.Sprintf("# 1 error occurred upgrading the following block:\n\t# %s\n", es[0])
			}

			points := make([]string, len(es))
			for i, err := range es {
				if i == len(es)-1 {
					points[i] = fmt.Sprintf("# %s", err)
					continue
				}
				points[i] = fmt.Sprintf("# %s\n", err)
			}

			return fmt.Sprintf(
				"# %d errors occurred upgrading the following block:\n\t%s",
				len(es), strings.Join(points, "\n\t"))
		},
	}

	funcMap := texttemplate.FuncMap{
		"aws_secretsmanager": func(a ...string) string {
			if len(a) == 2 {
				for key, config := range amazonSecretsManagerMap {
					nameOk := config["name"] == a[0]
					keyOk := config["key"] == a[1]
					if nameOk && keyOk {
						return fmt.Sprintf("${data.amazon-secretsmanager.%s.value}", key)
					}
				}
				id := fmt.Sprintf("autogenerated_%d", len(amazonSecretsManagerMap)+1)
				amazonSecretsManagerMap[id] = map[string]interface{}{
					"name": a[0],
					"key":  a[1],
				}
				return fmt.Sprintf("${data.amazon-secretsmanager.%s.value}", id)
			}
			for key, config := range amazonSecretsManagerMap {
				nameOk := config["name"] == a[0]
				if nameOk {
					return fmt.Sprintf("${data.amazon-secretsmanager.%s.value}", key)
				}
			}
			id := fmt.Sprintf("autogenerated_%d", len(amazonSecretsManagerMap)+1)
			amazonSecretsManagerMap[id] = map[string]interface{}{
				"name": a[0],
			}
			return fmt.Sprintf("${data.amazon-secretsmanager.%s.value}", id)
		},
		"timestamp": func() string {
			timestamp = true
			return "${local.timestamp}"
		},
		"isotime": func(a ...string) string {
			timestamp = true
			return "${local.timestamp}"
		},
		"user": func(in string) string {
			if _, ok := localsVariableMap[in]; ok {
				// variable is now a local
				return fmt.Sprintf("${local.%s}", in)
			}
			return fmt.Sprintf("${var.%s}", in)
		},
		"env": func(in string) string {
			return fmt.Sprintf("${env(%q)}", in)
		},
		"build": func(a string) string {
			return fmt.Sprintf("${build.%s}", a)
		},
		"data": func(a string) string {
			return fmt.Sprintf("${data.%s}", a)
		},
		"template_dir": func() string {
			return fmt.Sprintf("${path.root}")
		},
		"pwd": func() string {
			return fmt.Sprintf("${path.cwd}")
		},
		"packer_version": func() string {
			return fmt.Sprintf("${packer.version}")
		},
		"uuid": func() string {
			return fmt.Sprintf("${uuidv4()}")
		},
		"lower": func(a string) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"lower",
				"`lower(var.example)`",
				"https://www.packer.io/docs/templates/hcl_templates/functions/string/lower",
			})
			return fmt.Sprintf("{{ lower `%s` }}", a), nil
		},
		"upper": func(a string) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"upper",
				"`upper(var.example)`",
				"https://www.packer.io/docs/templates/hcl_templates/functions/string/upper",
			})
			return fmt.Sprintf("{{ upper `%s` }}", a), nil
		},
		"split": func(a, b string, n int) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"split",
				"`split(separator, string)`",
				"https://www.packer.io/docs/templates/hcl_templates/functions/string/split",
			})
			return fmt.Sprintf("{{ split `%s` `%s` %d }}", a, b, n), nil
		},
		"replace": func(a, b string, n int, c string) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"replace",
				"`replace(string, substring, replacement)` or `regex_replace(string, substring, replacement)`",
				"https://www.packer.io/docs/templates/hcl_templates/functions/string/replace or https://www.packer.io/docs/templates/hcl_templates/functions/string/regex_replace",
			})
			return fmt.Sprintf("{{ replace `%s` `%s` `%s` %d }}", a, b, c, n), nil
		},
		"replace_all": func(a, b, c string) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"replace_all",
				"`replace(string, substring, replacement)` or `regex_replace(string, substring, replacement)`",
				"https://www.packer.io/docs/templates/hcl_templates/functions/string/replace or https://www.packer.io/docs/templates/hcl_templates/functions/string/regex_replace",
			})
			return fmt.Sprintf("{{ replace_all `%s` `%s` `%s` }}", a, b, c), nil
		},
		"clean_resource_name": func(a string) (string, error) {
			funcErrors = multierror.Append(funcErrors, UnhandleableArgumentError{
				"clean_resource_name",
				"use custom validation rules, `replace(string, substring, replacement)` or `regex_replace(string, substring, replacement)`",
				"https://packer.io/docs/templates/hcl_templates/variables#custom-validation-rules" +
					" , https://www.packer.io/docs/templates/hcl_templates/functions/string/replace" +
					" or https://www.packer.io/docs/templates/hcl_templates/functions/string/regex_replace",
			})
			return fmt.Sprintf("{{ clean_resource_name `%s` }}", a), nil
		},
		"build_name": func() string {
			return fmt.Sprintf("${build.name}")
		},
		"build_type": func() string {
			return fmt.Sprintf("${build.type}")
		},
	}

	tpl, err := texttemplate.New("hcl2_upgrade").
		Funcs(funcMap).
		Parse(string(s))

	if err != nil {
		return fallbackReturn(err)
	}

	str := &bytes.Buffer{}
	// PASSTHROUGHS is a map of variable-specific golang text template fields
	// that should remain in the text template format.
	if err := tpl.Execute(str, PASSTHROUGHS); err != nil {
		return fallbackReturn(err)
	}

	out := str.Bytes()

	if funcErrors.Len() > 0 {
		return append([]byte(fmt.Sprintf("\n%s", funcErrors)), out...)
	}
	return out
}

// variableTransposeTemplatingCalls executes parts of blocks as go template files and replaces
// their result with their hcl2 variant for variables block only. If something goes wrong the template
// containing the go template string is returned.
// In variableTransposeTemplatingCalls the definition of aws_secretsmanager function will create a data source
// with the same name as the variable.
func variableTransposeTemplatingCalls(s []byte) (isLocal bool, body []byte) {
	fallbackReturn := func(err error) []byte {
		if strings.Contains(err.Error(), "unhandled") {
			return append([]byte(fmt.Sprintf("\n# %s\n", err)), s...)
		}

		return append([]byte(fmt.Sprintf("\n# could not parse template for following block: %q\n", err)), s...)
	}

	setIsLocal := func(a ...string) string {
		isLocal = true
		return ""
	}

	// Make locals from variables using valid template engine,
	// expect the ones using only 'env'
	// ref: https://www.packer.io/docs/templates/legacy_json_templates/engine#template-engine
	funcMap := texttemplate.FuncMap{
		"aws_secretsmanager": setIsLocal,
		"timestamp":          setIsLocal,
		"isotime":            setIsLocal,
		"user":               setIsLocal,
		"env": func(in string) string {
			return fmt.Sprintf("${env(%q)}", in)
		},
		"template_dir":   setIsLocal,
		"pwd":            setIsLocal,
		"packer_version": setIsLocal,
		"uuid":           setIsLocal,
		"lower":          setIsLocal,
		"upper":          setIsLocal,
		"split": func(_, _ string, _ int) (string, error) {
			isLocal = true
			return "", nil
		},
		"replace": func(_, _ string, _ int, _ string) (string, error) {
			isLocal = true
			return "", nil
		},
		"replace_all": func(_, _, _ string) (string, error) {
			isLocal = true
			return "", nil
		},
	}

	tpl, err := texttemplate.New("hcl2_upgrade").
		Funcs(funcMap).
		Parse(string(s))

	if err != nil {
		return isLocal, fallbackReturn(err)
	}

	str := &bytes.Buffer{}
	// PASSTHROUGHS is a map of variable-specific golang text template fields
	// that should remain in the text template format.
	if err := tpl.Execute(str, PASSTHROUGHS); err != nil {
		return isLocal, fallbackReturn(err)
	}

	return isLocal, str.Bytes()
}

func jsonBodyToHCL2Body(out *hclwrite.Body, kvs map[string]interface{}) {
	ks := []string{}
	for k := range kvs {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		value := kvs[k]

		switch value := value.(type) {
		case map[string]interface{}:
			var mostComplexElem interface{}
			for _, randomElem := range value {
				// HACK: we take the most complex element of that map because
				// in HCL2, map of objects can be bodies, for example:
				// map containing object: source_ami_filter {} ( body )
				// simple string/string map: tags = {} ) ( attribute )
				//
				// if we could not find an object in this map then it's most
				// likely a plain map and so we guess it should be and
				// attribute. Though now if value refers to something that is
				// an object but only contains a string or a bool; we could
				// generate a faulty object. For example a (somewhat invalid)
				// source_ami_filter where only `most_recent` is set.
				switch randomElem.(type) {
				case string, int, float64, bool:
					if mostComplexElem != nil {
						continue
					}
					mostComplexElem = randomElem
				default:
					mostComplexElem = randomElem
				}
			}

			switch mostComplexElem.(type) {
			case string, int, float64, bool:
				out.SetAttributeValue(k, hcl2shim.HCL2ValueFromConfigValue(value))
			default:
				nestedBlockBody := out.AppendNewBlock(k, nil).Body()
				jsonBodyToHCL2Body(nestedBlockBody, value)
			}
		case map[string]string, map[string]int, map[string]float64:
			out.SetAttributeValue(k, hcl2shim.HCL2ValueFromConfigValue(value))
		case []interface{}:
			if len(value) == 0 {
				continue
			}

			var mostComplexElem interface{}
			for _, randomElem := range value {
				// HACK: we take the most complex element of that slice because
				// in hcl2 slices of plain types can be arrays, for example:
				// simple string type: owners = ["0000000000"]
				// object: launch_block_device_mappings {}
				switch randomElem.(type) {
				case string, int, float64, bool:
					if mostComplexElem != nil {
						continue
					}
					mostComplexElem = randomElem
				default:
					mostComplexElem = randomElem
				}
			}
			switch mostComplexElem.(type) {
			case map[string]interface{}:
				// this is an object in a slice; so we unwrap it. We
				// could try to remove any 's' suffix in the key, but
				// this might not work everywhere.
				for i := range value {
					value := value[i].(map[string]interface{})
					nestedBlockBody := out.AppendNewBlock(k, nil).Body()
					jsonBodyToHCL2Body(nestedBlockBody, value)
				}
				continue
			default:
				out.SetAttributeValue(k, hcl2shim.HCL2ValueFromConfigValue(value))
			}
		default:
			out.SetAttributeValue(k, hcl2shim.HCL2ValueFromConfigValue(value))
		}
	}
}

func isSensitiveVariable(key string, vars []*template.Variable) bool {
	for _, v := range vars {
		if v.Key == key {
			return true
		}
	}
	return false
}

func (*HCL2UpgradeCommand) Help() string {
	helpText := `
Usage: packer hcl2_upgrade [options] TEMPLATE

  Will transform your JSON template into an HCL2 configuration.

Options:

  -output-file=path    Set output file name. By default this will be the
                       TEMPLATE name with ".pkr.hcl" appended to it. To be a
                       valid Packer HCL template, it must have the suffix
                       ".pkr.hcl"
  -with-annotations    Add helper annotation comments to the file to help new
                       HCL2 users understand the template format.
`

	return strings.TrimSpace(helpText)
}

func (*HCL2UpgradeCommand) Synopsis() string {
	return "transform a JSON template into an HCL2 configuration"
}

func (*HCL2UpgradeCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*HCL2UpgradeCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{}
}

// Specific blocks parser responsible to parse and write the block

type PackerParser struct {
	WithAnnotations bool
	out             []byte
}

func (p *PackerParser) Parse(tpl *template.Template) error {
	if tpl.MinVersion != "" {
		fileContent := hclwrite.NewEmptyFile()
		body := fileContent.Body()
		packerBody := body.AppendNewBlock("packer", nil).Body()
		packerBody.SetAttributeValue("required_version", cty.StringVal(fmt.Sprintf(">= %s", tpl.MinVersion)))
		p.out = fileContent.Bytes()
	}
	return nil
}

func (p *PackerParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(packerBlockHeader))
		}
		out.Write(p.out)
	}
}

type VariableParser struct {
	WithAnnotations bool
	variablesOut    []byte
	localsOut       []byte
}

func (p *VariableParser) Parse(tpl *template.Template) error {
	// OutPut Locals and Local blocks
	localsContent := hclwrite.NewEmptyFile()
	localsBody := localsContent.Body()
	localsBody.AppendNewline()
	localBody := localsBody.AppendNewBlock("locals", nil).Body()

	if len(p.variablesOut) == 0 {
		p.variablesOut = []byte{}
	}
	if len(p.localsOut) == 0 {
		p.localsOut = []byte{}
	}

	variables := []*template.Variable{}
	{
		// sort variables to avoid map's randomness
		for _, variable := range tpl.Variables {
			variables = append(variables, variable)
		}
		sort.Slice(variables, func(i, j int) bool {
			return variables[i].Key < variables[j].Key
		})
	}

	hasLocals := false
	for _, variable := range variables {
		variablesContent := hclwrite.NewEmptyFile()
		variablesBody := variablesContent.Body()
		variablesBody.AppendNewline()
		variableBody := variablesBody.AppendNewBlock("variable", []string{variable.Key}).Body()
		variableBody.SetAttributeRaw("type", hclwrite.Tokens{&hclwrite.Token{Bytes: []byte("string")}})

		if variable.Default != "" || !variable.Required {
			variableBody.SetAttributeValue("default", hcl2shim.HCL2ValueFromConfigValue(variable.Default))
		}
		sensitive := false
		if isSensitiveVariable(variable.Key, tpl.SensitiveVariables) {
			sensitive = true
			variableBody.SetAttributeValue("sensitive", cty.BoolVal(true))
		}
		isLocal, out := variableTransposeTemplatingCalls(variablesContent.Bytes())
		if isLocal {
			if sensitive {
				// Create Local block because this is sensitive
				localContent := hclwrite.NewEmptyFile()
				body := localContent.Body()
				body.AppendNewline()
				localBody := body.AppendNewBlock("local", []string{variable.Key}).Body()
				localBody.SetAttributeValue("sensitive", cty.BoolVal(true))
				localBody.SetAttributeValue("expression", hcl2shim.HCL2ValueFromConfigValue(variable.Default))
				p.localsOut = append(p.localsOut, transposeTemplatingCalls(localContent.Bytes())...)
				localsVariableMap[variable.Key] = "local"
				continue
			}
			localBody.SetAttributeValue(variable.Key, hcl2shim.HCL2ValueFromConfigValue(variable.Default))
			localsVariableMap[variable.Key] = "locals"
			hasLocals = true
			continue
		}
		p.variablesOut = append(p.variablesOut, out...)
	}

	if hasLocals {
		p.localsOut = append(p.localsOut, transposeTemplatingCalls(localsContent.Bytes())...)
	}
	return nil
}

func (p *VariableParser) Write(out *bytes.Buffer) {
	if len(p.variablesOut) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(inputVarHeader))
		}
		out.Write(p.variablesOut)
	}
}

type LocalsParser struct {
	WithAnnotations bool
	LocalsOut       []byte
}

func (p *LocalsParser) Parse(tpl *template.Template) error {
	// Locals where parsed with Variables
	return nil
}

func (p *LocalsParser) Write(out *bytes.Buffer) {
	if timestamp {
		_, _ = out.Write([]byte("\n"))
		if p.WithAnnotations {
			fmt.Fprintln(out, `# "timestamp" template function replacement`)
		}
		fmt.Fprintln(out, `locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }`)
	}
	if len(p.LocalsOut) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(localsVarHeader))
		}
		out.Write(p.LocalsOut)
	}
}

type AmazonSecretsDatasourceParser struct {
	WithAnnotations bool
	out             []byte
}

func (p *AmazonSecretsDatasourceParser) Parse(_ *template.Template) error {
	if p.out == nil {
		p.out = []byte{}
	}

	keys := make([]string, 0, len(amazonSecretsManagerMap))
	for k := range amazonSecretsManagerMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, dataSourceName := range keys {
		datasourceContent := hclwrite.NewEmptyFile()
		body := datasourceContent.Body()
		body.AppendNewline()
		datasourceBody := body.AppendNewBlock("data", []string{"amazon-secretsmanager", dataSourceName}).Body()
		jsonBodyToHCL2Body(datasourceBody, amazonSecretsManagerMap[dataSourceName])
		p.out = append(p.out, datasourceContent.Bytes()...)
	}

	return nil
}

func (p *AmazonSecretsDatasourceParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(amazonSecretsManagerDataHeader))
		}
		out.Write(p.out)
	}
}

type AmazonAmiDatasourceParser struct {
	Builders        []*template.Builder
	WithAnnotations bool
	out             []byte
}

func (p *AmazonAmiDatasourceParser) Parse(_ *template.Template) error {
	if p.out == nil {
		p.out = []byte{}
	}

	amazonAmiFilters := []map[string]interface{}{}
	i := 1
	for _, builder := range p.Builders {
		if strings.HasPrefix(builder.Type, "amazon-") {
			if sourceAmiFilter, ok := builder.Config["source_ami_filter"]; ok {
				sourceAmiFilterCfg := map[string]interface{}{}
				if err := mapstructure.Decode(sourceAmiFilter, &sourceAmiFilterCfg); err != nil {
					return fmt.Errorf("Failed to write amazon-ami data source: %v", err)
				}

				sourceAmiFilterCfg, err := copyAWSAccessConfig(sourceAmiFilterCfg, builder.Config)
				if err != nil {
					return err
				}

				duplicate := false
				dataSourceName := fmt.Sprintf("autogenerated_%d", i)
				for j, filter := range amazonAmiFilters {
					if reflect.DeepEqual(filter, sourceAmiFilterCfg) {
						duplicate = true
						dataSourceName = fmt.Sprintf("autogenerated_%d", j+1)
						continue
					}
				}

				// This is a hack...
				// Use templating so that it could be correctly transformed later into a data resource
				sourceAmiDataRef := fmt.Sprintf("{{ data `amazon-ami.%s.id` }}", dataSourceName)

				if duplicate {
					delete(builder.Config, "source_ami_filter")
					builder.Config["source_ami"] = sourceAmiDataRef
					continue
				}

				amazonAmiFilters = append(amazonAmiFilters, sourceAmiFilterCfg)
				delete(builder.Config, "source_ami_filter")
				builder.Config["source_ami"] = sourceAmiDataRef
				i++

				datasourceContent := hclwrite.NewEmptyFile()
				body := datasourceContent.Body()
				body.AppendNewline()
				sourceBody := body.AppendNewBlock("data", []string{"amazon-ami", dataSourceName}).Body()
				jsonBodyToHCL2Body(sourceBody, sourceAmiFilterCfg)
				p.out = append(p.out, transposeTemplatingCalls(datasourceContent.Bytes())...)
			}
		}
	}
	return nil
}

func copyAWSAccessConfig(sourceAmi map[string]interface{}, builder map[string]interface{}) (map[string]interface{}, error) {
	// Transform access config to a map
	accessConfigMap := map[string]interface{}{}
	if err := mapstructure.Decode(awscommon.AccessConfig{}, &accessConfigMap); err != nil {
		return sourceAmi, err
	}

	for k := range accessConfigMap {
		// Copy only access config present in the builder
		if v, ok := builder[k]; ok {
			sourceAmi[k] = v
		}
	}

	return sourceAmi, nil
}

func (p *AmazonAmiDatasourceParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(amazonAmiDataHeader))
		}
		out.Write(p.out)
	}
}

type SourceParser struct {
	Builders        []*template.Builder
	BuilderPlugins  packer.BuilderSet
	WithAnnotations bool
	out             []byte
}

func (p *SourceParser) Parse(tpl *template.Template) error {
	if p.out == nil {
		p.out = []byte{}
	}
	for i, builderCfg := range p.Builders {
		sourcesContent := hclwrite.NewEmptyFile()
		body := sourcesContent.Body()

		body.AppendNewline()
		if !p.BuilderPlugins.Has(builderCfg.Type) {
			return fmt.Errorf("unknown builder type: %q\n", builderCfg.Type)
		}
		if builderCfg.Name == "" || builderCfg.Name == builderCfg.Type {
			builderCfg.Name = fmt.Sprintf("autogenerated_%d", i+1)
		}
		builderCfg.Name = strings.ReplaceAll(strings.TrimSpace(builderCfg.Name), " ", "_")

		sourceBody := body.AppendNewBlock("source", []string{builderCfg.Type, builderCfg.Name}).Body()

		jsonBodyToHCL2Body(sourceBody, builderCfg.Config)

		p.out = append(p.out, transposeTemplatingCalls(sourcesContent.Bytes())...)
	}
	return nil
}

func (p *SourceParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(sourcesHeader))
		}
		out.Write(p.out)
	}
}

type BuildParser struct {
	Builders        []*template.Builder
	WithAnnotations bool

	provisioners   BlockParser
	postProcessors BlockParser
	out            []byte
}

func (p *BuildParser) Parse(tpl *template.Template) error {
	if len(p.Builders) == 0 {
		return nil
	}

	buildContent := hclwrite.NewEmptyFile()
	buildBody := buildContent.Body()
	if tpl.Description != "" {
		buildBody.SetAttributeValue("description", cty.StringVal(tpl.Description))
		buildBody.AppendNewline()
	}

	sourceNames := []string{}
	for _, builder := range p.Builders {
		sourceNames = append(sourceNames, fmt.Sprintf("source.%s.%s", builder.Type, builder.Name))
	}
	buildBody.SetAttributeValue("sources", hcl2shim.HCL2ValueFromConfigValue(sourceNames))
	buildBody.AppendNewline()
	p.out = buildContent.Bytes()

	p.provisioners = &ProvisionerParser{
		WithAnnotations: p.WithAnnotations,
	}
	if err := p.provisioners.Parse(tpl); err != nil {
		return err
	}

	p.postProcessors = &PostProcessorParser{
		WithAnnotations: p.WithAnnotations,
	}
	if err := p.postProcessors.Parse(tpl); err != nil {
		return err
	}

	return nil
}

func (p *BuildParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		if p.WithAnnotations {
			out.Write([]byte(buildHeader))
		} else {
			_, _ = out.Write([]byte("\n"))
		}
		_, _ = out.Write([]byte("build {\n"))
		out.Write(p.out)
		p.provisioners.Write(out)
		p.postProcessors.Write(out)
		_, _ = out.Write([]byte("}\n"))
	}
}

type ProvisionerParser struct {
	WithAnnotations bool
	out             []byte
}

func (p *ProvisionerParser) Parse(tpl *template.Template) error {
	if p.out == nil {
		p.out = []byte{}
	}
	for _, provisioner := range tpl.Provisioners {
		contentBytes := writeProvisioner("provisioner", provisioner)
		p.out = append(p.out, transposeTemplatingCalls(contentBytes)...)
	}

	if tpl.CleanupProvisioner != nil {
		contentBytes := writeProvisioner("error-cleanup-provisioner", tpl.CleanupProvisioner)
		p.out = append(p.out, transposeTemplatingCalls(contentBytes)...)
	}
	return nil
}

func writeProvisioner(typeName string, provisioner *template.Provisioner) []byte {
	provisionerContent := hclwrite.NewEmptyFile()
	body := provisionerContent.Body()
	block := body.AppendNewBlock(typeName, []string{provisioner.Type})
	cfg := provisioner.Config
	if len(provisioner.Except) > 0 {
		cfg["except"] = provisioner.Except
	}
	if len(provisioner.Only) > 0 {
		cfg["only"] = provisioner.Only
	}
	if provisioner.MaxRetries != "" {
		cfg["max_retries"] = provisioner.MaxRetries
	}
	if provisioner.Timeout > 0 {
		cfg["timeout"] = provisioner.Timeout.String()
	}
	if provisioner.PauseBefore > 0 {
		cfg["pause_before"] = provisioner.PauseBefore.String()
	}
	body.AppendNewline()
	jsonBodyToHCL2Body(block.Body(), cfg)
	return provisionerContent.Bytes()
}

func (p *ProvisionerParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		out.Write(p.out)
	}
}

type PostProcessorParser struct {
	WithAnnotations bool
	out             []byte
}

func (p *PostProcessorParser) Parse(tpl *template.Template) error {
	if p.out == nil {
		p.out = []byte{}
	}
	for _, pps := range tpl.PostProcessors {
		postProcessorContent := hclwrite.NewEmptyFile()
		body := postProcessorContent.Body()

		switch len(pps) {
		case 0:
			continue
		case 1:
		default:
			body = body.AppendNewBlock("post-processors", nil).Body()
		}
		for _, pp := range pps {
			ppBody := body.AppendNewBlock("post-processor", []string{pp.Type}).Body()
			if pp.KeepInputArtifact != nil {
				ppBody.SetAttributeValue("keep_input_artifact", cty.BoolVal(*pp.KeepInputArtifact))
			}
			cfg := pp.Config
			if len(pp.Except) > 0 {
				cfg["except"] = pp.Except
			}
			if len(pp.Only) > 0 {
				cfg["only"] = pp.Only
			}
			if pp.Name != "" && pp.Name != pp.Type {
				cfg["name"] = pp.Name
			}
			jsonBodyToHCL2Body(ppBody, cfg)
		}

		p.out = append(p.out, transposeTemplatingCalls(postProcessorContent.Bytes())...)
	}
	return nil
}

func (p *PostProcessorParser) Write(out *bytes.Buffer) {
	if len(p.out) > 0 {
		out.Write(p.out)
	}
}

var PASSTHROUGHS = map[string]string{"NVME_Present": "{{ .NVME_Present }}",
	"Usb_Present":                "{{ .Usb_Present }}",
	"Serial_Type":                "{{ .Serial_Type }}",
	"MapKey":                     "{{ .MapKey }}",
	"HostAlias":                  "{{ .HostAlias }}",
	"BoxName":                    "{{ .BoxName }}",
	"Port":                       "{{ .Port }}",
	"Header":                     "{{ .Header }}",
	"HTTPIP":                     "{{ .HTTPIP }}",
	"Host":                       "{{ .Host }}",
	"PACKER_TEST_TEMP":           "{{ .PACKER_TEST_TEMP }}",
	"SCSI_diskAdapterType":       "{{ .SCSI_diskAdapterType }}",
	"VHDBlockSizeBytes":          "{{ .VHDBlockSizeBytes }}",
	"Parallel_Auto":              "{{ .Parallel_Auto }}",
	"KTyp":                       "{{ .KTyp }}",
	"MemorySize":                 "{{ .MemorySize }}",
	"APIURL":                     "{{ .APIURL }}",
	"SourcePath":                 "{{ .SourcePath }}",
	"CDROMType":                  "{{ .CDROMType }}",
	"Parallel_Present":           "{{ .Parallel_Present }}",
	"HTTPPort":                   "{{ .HTTPPort }}",
	"BuildName":                  "{{ .BuildName }}",
	"Network_Device":             "{{ .Network_Device }}",
	"Flavor":                     "{{ .Flavor }}",
	"Image":                      "{{ .Image }}",
	"Os":                         "{{ .Os }}",
	"Network_Type":               "{{ .Network_Type }}",
	"SourceOMIName":              "{{ .SourceOMIName }}",
	"Serial_Yield":               "{{ .Serial_Yield }}",
	"SourceAMI":                  "{{ .SourceAMI }}",
	"SSHHostPort":                "{{ .SSHHostPort }}",
	"Vars":                       "{{ .Vars }}",
	"Slice":                      "{{ .Slice }}",
	"Version":                    "{{ .Version }}",
	"Parallel_Bidirectional":     "{{ .Parallel_Bidirectional }}",
	"Serial_Auto":                "{{ .Serial_Auto }}",
	"VHDX":                       "{{ .VHDX }}",
	"WinRMPassword":              "{{ .WinRMPassword }}",
	"DefaultOrganizationID":      "{{ .DefaultOrganizationID }}",
	"HTTPDir":                    "{{ .HTTPDir }}",
	"SegmentPath":                "{{ .SegmentPath }}",
	"NewVHDSizeBytes":            "{{ .NewVHDSizeBytes }}",
	"CTyp":                       "{{ .CTyp }}",
	"VMName":                     "{{ .VMName }}",
	"Serial_Present":             "{{ .Serial_Present }}",
	"Varname":                    "{{ .Varname }}",
	"DiskNumber":                 "{{ .DiskNumber }}",
	"SecondID":                   "{{ .SecondID }}",
	"Typ":                        "{{ .Typ }}",
	"SourceAMIName":              "{{ .SourceAMIName }}",
	"ActiveProfile":              "{{ .ActiveProfile }}",
	"Primitive":                  "{{ .Primitive }}",
	"Elem":                       "{{ .Elem }}",
	"Network_Adapter":            "{{ .Network_Adapter }}",
	"Minor":                      "{{ .Minor }}",
	"ProjectName":                "{{ .ProjectName }}",
	"Generation":                 "{{ .Generation }}",
	"User":                       "{{ .User }}",
	"Size":                       "{{ .Size }}",
	"Parallel_Filename":          "{{ .Parallel_Filename }}",
	"ID":                         "{{ .ID }}",
	"FastpathLen":                "{{ .FastpathLen }}",
	"Tag":                        "{{ .Tag }}",
	"Serial_Endpoint":            "{{ .Serial_Endpoint }}",
	"GuestOS":                    "{{ .GuestOS }}",
	"Major":                      "{{ .Major }}",
	"Serial_Filename":            "{{ .Serial_Filename }}",
	"Name":                       "{{ .Name }}",
	"SourceOMI":                  "{{ .SourceOMI }}",
	"SCSI_Present":               "{{ .SCSI_Present }}",
	"CpuCount":                   "{{ .CpuCount }}",
	"DefaultProjectID":           "{{ .DefaultProjectID }}",
	"CDROMType_PrimarySecondary": "{{ .CDROMType_PrimarySecondary }}",
	"Arch":                       "{{ .Arch }}",
	"ImageFile":                  "{{ .ImageFile }}",
	"SATA_Present":               "{{ .SATA_Present }}",
	"Serial_Host":                "{{ .Serial_Host }}",
	"BuildRegion":                "{{ .BuildRegion }}",
	"Id":                         "{{ .Id }}",
	"SyncedFolder":               "{{ .SyncedFolder }}",
	"Network_Name":               "{{ .Network_Name }}",
	"AccountID":                  "{{ .AccountID }}",
	"OPTION":                     "{{ .OPTION }}",
	"Type":                       "{{ .Type }}",
	"CustomVagrantfile":          "{{ .CustomVagrantfile }}",
	"SendTelemetry":              "{{ .SendTelemetry }}",
	"DiskType":                   "{{ .DiskType }}",
	"Password":                   "{{ .Password }}",
	"HardDrivePath":              "{{ .HardDrivePath }}",
	"ISOPath":                    "{{ .ISOPath }}",
	"Insecure":                   "{{ .Insecure }}",
	"Region":                     "{{ .Region }}",
	"SecretKey":                  "{{ .SecretKey }}",
	"DefaultRegion":              "{{ .DefaultRegion }}",
	"MemoryStartupBytes":         "{{ .MemoryStartupBytes }}",
	"SwitchName":                 "{{ .SwitchName }}",
	"Path":                       "{{ .Path }}",
	"Username":                   "{{ .Username }}",
	"OutputDir":                  "{{ .OutputDir }}",
	"DiskName":                   "{{ .DiskName }}",
	"ProviderVagrantfile":        "{{ .ProviderVagrantfile }}",
	"Sound_Present":              "{{ .Sound_Present }}",
}
