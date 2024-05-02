// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"runtime"
	"strings"

	gversion "github.com/hashicorp/go-version"
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
	flags := c.Meta.FlagSet("init")
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

	if len(reqs) == 0 {
		c.Ui.Message(`
No plugins requirement found, make sure you reference a Packer config
containing a packer.required_plugins block. See
https://www.packer.io/docs/templates/hcl_templates/blocks/packer
for more info.`)
	}

	opts := plugingetter.ListInstallationsOptions{
		PluginDirectory: c.Meta.CoreConfig.Components.PluginConfig.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
			ReleasesOnly: true,
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

		if len(installs) > 0 {
			if !cla.Force && !cla.Upgrade {
				continue
			}

			if cla.Force && !cla.Upgrade {
				// Only place another constaint to the latest release
				// binary, if any, otherwise this is essentially the same
				// as an upgrade
				var installVersion string
				for _, install := range installs {
					ver, _ := gversion.NewVersion(install.Version)
					if ver.Prerelease() == "" {
						installVersion = install.Version
					}
				}

				if installVersion != "" {
					pluginRequirement.VersionConstraints, _ = gversion.NewConstraint(fmt.Sprintf("=%s", installVersion))
				}
			}
		}

		newInstall, err := pluginRequirement.InstallLatest(plugingetter.InstallOptions{
			PluginDirectory:           opts.PluginDirectory,
			BinaryInstallationOptions: opts.BinaryInstallationOptions,
			Getters:                   getters,
			Force:                     cla.Force,
		})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed getting the %q plugin:", pluginRequirement.Identifier))
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
Usage: packer init [options] TEMPLATE

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
  -force                       Forces reinstallation of plugins, even if already
                               installed.
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
