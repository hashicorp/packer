package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/packer/template"
	"github.com/posener/complete"
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
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	args = flags.Args()
	if len(args) != 2 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.Path = args[0]
	cfg.OutputDir = args[1]
	return &cfg, 0
}

func (c *HCL2UpgradeCommand) RunContext(buildCtx context.Context, cla *HCL2UpgradeArgs) int {

	if err := os.MkdirAll(cla.OutputDir, 0); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to create output directory: %v", err))
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

	variablesOutput := &strings.Builder{}
	for _, variable := range variables {
		fmt.Fprintf(variablesOutput, "\nvariable %q {\n", variable.Key)
		if variable.Default != "" || !variable.Required {
			fmt.Fprintf(variablesOutput, "  default = %q\n", variable.Default)
		}
		if isSensitiveVariable(variable.Key, tpl.SensitiveVariables) {
			fmt.Fprintln(variablesOutput, "  sensitive = true")
		}
		fmt.Fprintln(variablesOutput, "}")
	}
	c.Ui.Say(variablesOutput.String())

	builders := []*template.Builder{}
	{
		// sort builders to avoid map's randomness
		for _, builder := range tpl.Builders {
			builders = append(builders, builder)
		}
		sort.Slice(builders, func(i, j int) bool {
			return builders[i].Type+builders[i].Name < builders[j].Type+builders[j].Name
		})
	}

	sourcesOutput := &strings.Builder{}
	for i, builderCfg := range builders {
		if !c.Meta.CoreConfig.Components.BuilderStore.Has(builderCfg.Type) {
			c.Ui.Error(fmt.Sprintf("unknown builder type: %q\n", builderCfg.Type))
			return 1
		}
		if builderCfg.Name == "" || builderCfg.Name == builderCfg.Type {
			builderCfg.Name = fmt.Sprintf("%d", i+1)
		}

		fmt.Fprintf(sourcesOutput, "\nsource %q %q {\n", builderCfg.Type, builderCfg.Name)
		fmt.Fprintln(sourcesOutput, "}")

	}
	c.Ui.Say(sourcesOutput.String())

	buildOutput := &strings.Builder{}
	fmt.Fprintf(buildOutput, "build {\n")
	for _, builder := range builders {
		if tpl.Description != "" {
			fmt.Fprintf(buildOutput, "\n  description = %q\n", tpl.Description)
		}
		fmt.Fprintf(buildOutput, "\n  sources = [")
		fmt.Fprintf(buildOutput, "\n    \"source.%s.%s\",\n", builder.Type, builder.Name)
		fmt.Fprintln(buildOutput, "  ]")
	}

	for _, provisioner := range tpl.Provisioners {
		fmt.Fprintf(buildOutput, "\n  provisioner %q {\n", provisioner.Type)

		fmt.Fprintln(buildOutput, "  }")
	}

	fmt.Fprintf(buildOutput, "}\n")
	c.Ui.Say(buildOutput.String())
	return 0
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
Usage: packer hcl2_upgrade JSON_TEMPLATE OUTPUT_DIR

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
