package command

import (
	"context"
	"crypto/sha256"
	"log"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/plugin"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/posener/complete"
)

type InitCommand struct {
	Meta
}

func (c *InitCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *InitCommand) ParseArgs(args []string) (*InitArgs, int) {
	var cfg InitArgs
	flags := c.Meta.FlagSet("init", 0)
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
	return &cfg, 0
}

func (c *InitCommand) RunContext(buildCtx context.Context, cla *InitArgs) int {
	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	// Get plugins requirements
	reqs, diags := packerStarter.PluginRequirements()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	opts := plugingetter.ListInstallationsOptions{
		FromFolders: c.Meta.CoreConfig.Components.KnownPluginFolders,
		OS:          runtime.GOOS,
		ARCH:        runtime.GOARCH,
		Extension:   plugin.FileExtension,
		Checksummers: []plugingetter.Checksummer{
			{Type: "sha256", Hash: sha256.New()},
		},
	}

	log.Printf("[TRACE] init: %#v", opts)

	for _, pluginRequirement := range reqs {
		// Get installed plugins that match requirement

		installs, err := pluginRequirement.ListInstallations(opts)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}

		log.Printf("[TRACE] for plugin %s found %d matching installation(s)", pluginRequirement.Identifier.String(), len(installs))

		if len(installs) > 0 && cla.Upgrade == false {
			continue
		}

		newInstall, err := pluginRequirement.InstallLatest(plugingetter.InstallLatestOptions{})
		if err != nil {
			c.Ui.Error(err.Error())
		}
		_ = newInstall

	}
	return ret
}

func (*InitCommand) Help() string {
	helpText := `
Usage: packer init [options] TEMPLATE

  TODO
  TODO

Options:

  -upgrade=false          Do you want to try upgrading the plugin if it is already present ? (default: false)
`

	return strings.TrimSpace(helpText)
}

func (*InitCommand) Synopsis() string {
	return "Install missing plugins or upgrade plugins"
}

func (*InitCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*InitCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-upgrade": complete.PredictNothing,
	}
}
