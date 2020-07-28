package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	texttemplate "text/template"

	"github.com/hashicorp/hcl/v2/hclwrite"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/template"
	"github.com/posener/complete"
	"github.com/zclconf/go-cty/cty"
)

type HCL2UpgradeCommand struct {
	Meta
	OutputDir string
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

func (c *HCL2UpgradeCommand) RunContext(buildCtx context.Context, cla *HCL2UpgradeArgs) int {

	out := &bytes.Buffer{}
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

	hdl, ret := c.GetConfigFromJSON(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	core := hdl.(*CoreWrapper).Core
	if err := core.Initialize(); err != nil {
		c.Ui.Error(fmt.Sprintf("Initialization error: %v", err))
		return 1
	}
	tpl := core.Template

	// Output variables section

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

	for _, variable := range variables {
		variablesContent := hclwrite.NewEmptyFile()
		variablesBody := variablesContent.Body()

		variableBody := variablesBody.AppendNewBlock("variable", []string{variable.Key}).Body()

		if variable.Default != "" || !variable.Required {
			variableBody.SetAttributeValue("default", hcl2shim.HCL2ValueFromConfigValue(variable.Default))
		}
		if isSensitiveVariable(variable.Key, tpl.SensitiveVariables) {
			variableBody.SetAttributeValue("sensitive", cty.BoolVal(true))
		}
		variablesBody.AppendNewline()
		out.Write(magicTemplate(variablesContent.Bytes()))
	}

	fmt.Fprintln(out, `# "timestamp" template function replacement`)
	fmt.Fprintln(out, `locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }`)

	// Output sources section

	builders := []*template.Builder{}
	{
		// sort builders to avoid map's randomnes
		for _, builder := range tpl.Builders {
			builders = append(builders, builder)
		}
		sort.Slice(builders, func(i, j int) bool {
			return builders[i].Type+builders[i].Name < builders[j].Type+builders[j].Name
		})
	}

	for i, builderCfg := range builders {
		sourcesContent := hclwrite.NewEmptyFile()
		body := sourcesContent.Body()

		body.AppendNewline()
		if !c.Meta.CoreConfig.Components.BuilderStore.Has(builderCfg.Type) {
			c.Ui.Error(fmt.Sprintf("unknown builder type: %q\n", builderCfg.Type))
			return 1
		}
		if builderCfg.Name == "" || builderCfg.Name == builderCfg.Type {
			builderCfg.Name = fmt.Sprintf("%d", i+1)
		}
		sourceBody := body.AppendNewBlock("source", []string{builderCfg.Type, builderCfg.Name}).Body()

		jsonBodyToHCL2Body(sourceBody, builderCfg.Config)

		_, _ = out.Write(magicTemplate(sourcesContent.Bytes()))
	}

	// Output build section

	_, _ = out.Write([]byte("\nbuild {\n"))

	buildContent := hclwrite.NewEmptyFile()
	buildBody := buildContent.Body()
	if tpl.Description != "" {
		buildBody.SetAttributeValue("description", cty.StringVal(tpl.Description))
		buildBody.AppendNewline()
	}

	sourceNames := []string{}
	for _, builder := range builders {
		sourceNames = append(sourceNames, fmt.Sprintf("source.%s.%s", builder.Type, builder.Name))
	}
	buildBody.SetAttributeValue("sources", hcl2shim.HCL2ValueFromConfigValue(sourceNames))
	buildBody.AppendNewline()
	_, _ = buildContent.WriteTo(out)

	for _, provisioner := range tpl.Provisioners {
		provisionerContent := hclwrite.NewEmptyFile()
		body := provisionerContent.Body()

		buildBody.AppendNewline()
		block := body.AppendNewBlock("provisioner", []string{provisioner.Type})
		jsonBodyToHCL2Body(block.Body(), provisioner.Config)

		out.Write(magicTemplate(provisionerContent.Bytes()))
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
			jsonBodyToHCL2Body(ppBody, pp.Config)
		}

		_, _ = out.Write(magicTemplate(postProcessorContent.Bytes()))
	}

	_, _ = out.Write([]byte("}\n"))

	_, _ = output.Write(hclwrite.Format(out.Bytes()))

	c.Ui.Say(fmt.Sprintf("Successfully created %s ", cla.OutputFile))

	return 0
}

// magicTemplate executes parts of blocks as go template files and replaces
// their result with their hcl2 variant. If something goes wrong the template
// containing the go template string is returned.
func magicTemplate(s []byte) []byte {
	fallbackReturn := func(err error) []byte {
		return append([]byte(fmt.Sprintf("#could not parse template for following block: %q\n", err)), s...)
	}
	funcMap := texttemplate.FuncMap{
		"isotime": func(string) string {
			return "${local.timestamp}"
		},
		"timestamp": func() string {
			return "${local.timestamp}"
		},
		"user": func(in string) string {
			return fmt.Sprintf("${var.%s}", in)
		},
		"env": func(in string) string {
			return fmt.Sprintf("${var.%s}", in)
		},
		"vault_key": func(a, b string) string {
			return fmt.Sprintf("{{ consul_key `%s` `%s` }}", a, b)
		},
		"build": func(a string) string {
			return fmt.Sprintf("${build.%s}", a)
		},
	}
	transparentFuncs := []string{
		"consul_key",
		"aws_secretsmanager",
	}
	for i := range transparentFuncs {
		v := transparentFuncs[i]
		funcMap[v] = func(in string) string {
			return fmt.Sprintf("{{ %s `%s` }}", v, in)
		}
	}

	tpl, err := texttemplate.New("generated").
		Funcs(funcMap).
		Parse(string(s))

	if err != nil {
		return fallbackReturn(err)
	}

	str := &bytes.Buffer{}
	v := struct {
		HTTPIP   string
		HTTPPort string
	}{
		HTTPIP:   "{{ .HTTPIP }}",
		HTTPPort: "{{ .HTTPPort }}",
	}
	if err := tpl.Execute(str, v); err != nil {
		return fallbackReturn(err)
	}

	return str.Bytes()
}

func jsonBodyToHCL2Body(out *hclwrite.Body, kvs map[string]interface{}) {
	ks := []string{}
	for k := range kvs {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		value := kvs[k]
		out.SetAttributeValue(k, hcl2shim.HCL2ValueFromConfigValue(value))
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
Usage: packer hcl2_upgrade -output-file=JSON_TEMPLATE.pkr.hcl JSON_TEMPLATE...

  Will transform your JSON template to a HCL2 configuration.
`

	return strings.TrimSpace(helpText)
}

func (*HCL2UpgradeCommand) Synopsis() string {
	return "build image(s) from template"
}

func (*HCL2UpgradeCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*HCL2UpgradeCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{}
}
