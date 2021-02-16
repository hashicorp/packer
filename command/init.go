package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"runtime"
	"strings"

	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/hashicorp/packer/packer/plugin-getter/github"
	"github.com/hashicorp/packer/version"
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
		FromFolders: c.Meta.CoreConfig.Components.PluginConfig.KnownPluginFolders,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	log.Printf("[TRACE] init: %#v", opts)

	getters := []plugingetter.Getter{
		&github.Getter{
			// In the past some terraform plugins downloads were blocked from a
			// specific aws region by s3. Changing the user agent unblocked the
			// downloads so having one user agent per version will help mitigate
			// that a little more. Especially in the case someone forks this
			// code to make it more aggressive or something.
			// TODO: allow to set this from the config file or an environment
			// variable.
			UserAgent: "packer-getter-github-" + version.String(),
		},
	}

	ui := &packer.ColoredUi{
		Color: packer.UiColorCyan,
		Ui:    c.Ui,
	}

	for _, pluginRequirement := range reqs {
		// Get installed plugins that match requirement

		installs, err := pluginRequirement.ListInstallations(opts)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}

		log.Printf("[TRACE] for plugin %s found %d matching installation(s)", pluginRequirement.Identifier, len(installs))

		if len(installs) > 0 && cla.Upgrade == false {
			continue
		}

		newInstall, err := pluginRequirement.InstallLatest(plugingetter.InstallOptions{
			InFolders:                 opts.FromFolders,
			BinaryInstallationOptions: opts.BinaryInstallationOptions,
			Getters:                   getters,
		})
		if err != nil {
			c.Ui.Error(err.Error())
			ret = 1
		}
		if newInstall != nil {
			msg := fmt.Sprintf("Installed plugin %s %s in %q", pluginRequirement.Identifier, newInstall.Version, newInstall.BinaryPath)
			ui.Say(msg)
		}
	}
	return ret
}

func (*InitCommand) Help() string {
	helpText := `
Usage: packer init [options] [config.pkr.hcl|folder/]

  Install all the missing plugins required in a Packer config. Note that Packer
  does not have a state.

  This is the first command that should be executed when working with a new
  or existing template.

  This command is always safe to run multiple times. Though subsequent runs may
  give errors, this command will never delete anything.

Options:
  -upgrade                     On top of installing missing plugins, update
                               installed plugins to the latest available
                               version, if there is a new higher one. Note that
                               this still takes into consideration the version
                               constraint of the config.
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
