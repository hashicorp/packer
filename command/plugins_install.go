// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/hashicorp/packer/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/hashicorp/packer/packer/plugin-getter/github"
	pkrversion "github.com/hashicorp/packer/version"
)

type PluginsInstallCommand struct {
	Meta
}

func (c *PluginsInstallCommand) Synopsis() string {
	return "Install latest Packer plugin [matching version constraint]"
}

func (c *PluginsInstallCommand) Help() string {
	helpText := `
Usage: packer plugins install [OPTIONS...] <plugin> [<version constraint>]

  This command will install the most recent compatible Packer plugin matching
  version constraint.
  When the version constraint is omitted, the most recent version will be
  installed.

  Ex: packer plugins install github.com/hashicorp/happycloud v1.2.3

Options:
  - path <path>: install the plugin from a locally-sourced plugin binary. This
                 installs the plugin where a normal invocation would, but will
	         not try to download it from a web server, but instead directly
	         install the binary for Packer to be able to load it later on.
  - force:       forces installation of a plugin, even if it is already there.
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsInstallCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cmdArgs, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cmdArgs)
}

type PluginsInstallArgs struct {
	MetaArgs
	PluginName string
	PluginPath string
	Version    string
	Force      bool
}

func (pa *PluginsInstallArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.StringVar(&pa.PluginPath, "path", "", "install the plugin from a specific path")
	flags.BoolVar(&pa.Force, "force", false, "force installation of a plugin, even if already installed")
	pa.MetaArgs.AddFlagSets(flags)
}

func (c *PluginsInstallCommand) ParseArgs(args []string) (*PluginsInstallArgs, int) {
	pa := &PluginsInstallArgs{}

	flags := c.Meta.FlagSet("plugins install")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	pa.AddFlagSets(flags)
	err := flags.Parse(args)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse options: %s", err))
		return pa, 1
	}

	args = flags.Args()
	if len(args) < 1 || len(args) > 2 {
		c.Ui.Error(fmt.Sprintf("Invalid arguments, expected either 1 or 2 positional arguments, got %d", len(args)))
		flags.Usage()
		return pa, 1
	}

	if len(args) == 2 {
		pa.Version = args[1]
	}

	pa.PluginName = args[0]

	return pa, 0
}

func (c *PluginsInstallCommand) RunContext(buildCtx context.Context, args *PluginsInstallArgs) int {
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

	plugin, diags := addrs.ParsePluginSourceString(args.PluginName)
	if diags.HasErrors() {
		c.Ui.Error(diags.Error())
		return 1
	}

	// If we did specify a binary to install the plugin from, we ignore
	// the Github-based getter in favour of installing it directly.
	if args.PluginPath != "" {
		return c.InstallFromBinary(args)
	}

	// a plugin requirement that matches them all
	pluginRequirement := plugingetter.Requirement{
		Identifier: plugin,
	}

	if args.Version != "" {
		constraints, err := version.NewConstraint(args.Version)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		pluginRequirement.VersionConstraints = constraints
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	getters := []plugingetter.Getter{
		&github.Getter{
			// In the past some terraform plugins downloads were blocked from a
			// specific aws region by s3. Changing the user agent unblocked the
			// downloads so having one user agent per version will help mitigate
			// that a little more. Especially in the case someone forks this
			// code to make it more aggressive or something.
			// TODO: allow to set this from the config file or an environment
			// variable.
			UserAgent: "packer-getter-github-" + pkrversion.String(),
		},
	}

	newInstall, err := pluginRequirement.InstallLatest(plugingetter.InstallOptions{
		InFolders:                 opts.FromFolders,
		BinaryInstallationOptions: opts.BinaryInstallationOptions,
		Getters:                   getters,
		Force:                     args.Force,
	})

	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	if newInstall != nil {
		msg := fmt.Sprintf("Installed plugin %s %s in %q", pluginRequirement.Identifier, newInstall.Version, newInstall.BinaryPath)
		ui := &packer.ColoredUi{
			Color: packer.UiColorCyan,
			Ui:    c.Ui,
		}
		ui.Say(msg)
		return 0
	}

	return 0
}

func (c *PluginsInstallCommand) InstallFromBinary(args *PluginsInstallArgs) int {
	pluginDirs := c.Meta.CoreConfig.Components.PluginConfig.KnownPluginFolders

	if len(pluginDirs) == 0 {
		c.Ui.Say(`Error: cannot find a place to install the plugin to

In order to install the plugin for later use, Packer needs to know where to
install them.

This can be specified through the PACKER_CONFIG_DIR environment variable,
but should be automatically inferred by Packer.

If you see this message, this is likely a Packer bug, please consider opening
an issue on our Github repo to signal it.`)
	}

	pluginSlugParts := strings.Split(args.PluginName, "/")
	if len(pluginSlugParts) != 3 {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin name specifier",
			Detail:   fmt.Sprintf("The plugin name specified provided (%q) does not conform to the mandated format of <host>/<org>/<plugin-name>.", args.PluginName),
		}})
	}

	// As with the other commands, we get the last plugin directory as it
	// has precedence over the others, and is where we'll install the
	// plugins to.
	pluginDir := pluginDirs[len(pluginDirs)-1]

	s, err := os.Stat(args.PluginPath)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to find plugin to promote",
			Detail:   fmt.Sprintf("The plugin %q failed to be opened because of an error: %s", args.PluginName, err),
		}})
	}

	if s.IsDir() {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Plugin to promote cannot be a directory",
			Detail:   "The packer plugin promote command can only install binaries, not directories",
		}})
	}

	describeCmd, err := exec.Command(args.PluginPath, "describe").Output()
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to describe the plugin",
			Detail:   fmt.Sprintf("Packer failed to run %s describe: %s", args.PluginPath, err),
		}})
	}
	var desc plugin.SetDescription
	if err := json.Unmarshal(describeCmd, &desc); err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to decode plugin describe info",
			Detail:   fmt.Sprintf("'%s describe' produced information that Packer couldn't decode: %s", args.PluginPath, err),
		}})
	}

	// Let's override the plugin's version if we specify it in the options
	// of the command
	if args.Version != "" {
		desc.Version = args.Version
	}

	pluginBinary, err := os.Open(args.PluginPath)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to open plugin binary",
			Detail:   fmt.Sprintf("Failed to open plugin binary from %q: %s", args.PluginPath, err),
		}})
	}
	defer pluginBinary.Close()

	// We'll install the SHA256SUM file alongside the plugin, based on the
	// contents of the plugin being passed.
	//
	// This will make our loaders happy as they require a valid checksum
	// for loading plugins installed this way.
	shasum := sha256.New()
	_, err = io.Copy(shasum, pluginBinary)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read plugin binary's contents",
			Detail:   fmt.Sprintf("Failed to read plugin binary from %q: %s", args.PluginPath, err),
		}})
	}

	// At this point, we know the provided binary behaves correctly with
	// describe, so it's very likely to be a plugin, let's install it.
	installDir := fmt.Sprintf("%s/%s", pluginDir, args.PluginName)
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to create output directory",
			Detail:   fmt.Sprintf("The installation directory %q failed to be created because of an error: %s", installDir, err),
		}})
	}

	binaryPath := fmt.Sprintf(
		"%s/packer-plugin-%s_v%s_%s_%s_%s",
		installDir,
		pluginSlugParts[2],
		desc.Version,
		desc.APIVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
	outputPlugin, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to create plugin binary",
			Detail:   fmt.Sprintf("Failed to create plugin binary at %q: %s", binaryPath, err),
		}})
	}
	defer outputPlugin.Close()

	_, err = pluginBinary.Seek(0, 0)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to reset plugin's reader",
			Detail:   fmt.Sprintf("Failed to seek offset 0 while attempting to reset the buffer for the plugin to install: %s", err),
		}})
	}

	_, err = io.Copy(outputPlugin, pluginBinary)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to copy plugin binary's contents",
			Detail:   fmt.Sprintf("Failed to copy plugin binary from %q to %q: %s", args.PluginPath, binaryPath, err),
		}})
	}

	shasumPath := fmt.Sprintf("%s_SHA256SUM", binaryPath)
	shaFile, err := os.OpenFile(shasumPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to create plugin SHA256SUM file",
			Detail:   fmt.Sprintf("Failed to create SHA256SUM file at %q: %s", shasumPath, err),
		}})
	}
	defer shaFile.Close()

	fmt.Fprintf(shaFile, "%x", shasum.Sum([]byte{}))

	c.Ui.Say(fmt.Sprintf("Successfully installed plugin %s from %s to %s", args.PluginName, args.PluginPath, binaryPath))

	return 0
}
