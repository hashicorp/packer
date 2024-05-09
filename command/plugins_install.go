// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
      packer plugins install --path ./packer-plugin-happycloud "github.com/hashicorp/happycloud"

Options:
  -path <path>                  Install the plugin from a locally-sourced plugin binary.
                                This installs the plugin where a normal invocation would, but will
                                not try to download it from a remote location, and instead
                                install the binary in the Packer plugins path. This option cannot
                                be specified with a version constraint.
  -force                        Forces reinstallation of plugins, even if already installed.
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
	PluginIdentifier string
	PluginPath       string
	Version          string
	Force            bool
}

func (pa *PluginsInstallArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.StringVar(&pa.PluginPath, "path", "", "install the binary specified by path as a Packer plugin.")
	flags.BoolVar(&pa.Force, "force", false, "force installation of the specified plugin, even if already installed.")
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

	if pa.Path != "" && pa.Version != "" {
		c.Ui.Error("Invalid arguments: a version cannot be specified when using --path to install a local plugin binary")
		flags.Usage()
		return pa, 1
	}

	pa.PluginIdentifier = args[0]
	return pa, 0
}

func (c *PluginsInstallCommand) RunContext(buildCtx context.Context, args *PluginsInstallArgs) int {
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
	if runtime.GOOS == "windows" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	plugin, err := addrs.ParsePluginSourceString(args.PluginIdentifier)
	if err != nil {
		c.Ui.Errorf("Invalid source string %q: %s", args.PluginIdentifier, err)
		return 1
	}

	// If we did specify a binary to install the plugin from, we ignore
	// the Github-based getter in favour of installing it directly.
	if args.PluginPath != "" {
		return c.InstallFromBinary(opts, plugin, args)
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
		PluginDirectory:           opts.PluginDirectory,
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

func (c *PluginsInstallCommand) InstallFromBinary(opts plugingetter.ListInstallationsOptions, pluginIdentifier *addrs.Plugin, args *PluginsInstallArgs) int {
	pluginDir := opts.PluginDirectory

	var err error

	args.PluginPath, err = filepath.Abs(args.PluginPath)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to transform path",
			Detail:   fmt.Sprintf("Failed to transform the given path to an absolute one: %s", err),
		}})
	}

	s, err := os.Stat(args.PluginPath)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to find plugin to promote",
			Detail:   fmt.Sprintf("The plugin %q failed to be opened because of an error: %s", args.PluginIdentifier, err),
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

	semver, err := version.NewSemver(desc.Version)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version",
			Detail:   fmt.Sprintf("Plugin's reported version (%q) is not semver-compatible: %s", desc.Version, err),
		}})
	}
	if semver.Prerelease() != "" && semver.Prerelease() != "dev" {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version",
			Detail:   fmt.Sprintf("Packer can only install plugin releases with this command (ex: 1.0.0) or development pre-releases (ex: 1.0.0-dev), the binary's reported version is %q", desc.Version),
		}})
	}

	pluginBinary, err := os.Open(args.PluginPath)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to open plugin binary",
			Detail:   fmt.Sprintf("Failed to open plugin binary from %q: %s", args.PluginPath, err),
		}})
	}

	pluginContents := bytes.Buffer{}
	_, err = io.Copy(&pluginContents, pluginBinary)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read plugin binary's contents",
			Detail:   fmt.Sprintf("Failed to read plugin binary from %q: %s", args.PluginPath, err),
		}})
	}
	_ = pluginBinary.Close()

	// At this point, we know the provided binary behaves correctly with
	// describe, so it's very likely to be a plugin, let's install it.
	installDir := filepath.Join(
		pluginDir,
		filepath.Join(pluginIdentifier.Parts()...),
	)
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to create output directory",
			Detail:   fmt.Sprintf("The installation directory %q failed to be created because of an error: %s", installDir, err),
		}})
	}

	// Remove metadata from plugin path
	noMetaVersion := semver.Core().String()
	if semver.Prerelease() != "" {
		noMetaVersion = fmt.Sprintf("%s-%s", noMetaVersion, semver.Prerelease())
	}

	outputPrefix := fmt.Sprintf(
		"packer-plugin-%s_v%s_%s",
		pluginIdentifier.Name(),
		noMetaVersion,
		desc.APIVersion,
	)
	binaryPath := filepath.Join(
		installDir,
		outputPrefix+opts.BinaryInstallationOptions.FilenameSuffix(),
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

	_, err = outputPlugin.Write(pluginContents.Bytes())
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to copy plugin binary's contents",
			Detail:   fmt.Sprintf("Failed to copy plugin binary from %q to %q: %s", args.PluginPath, binaryPath, err),
		}})
	}

	// We'll install the SHA256SUM file alongside the plugin, based on the
	// contents of the plugin being passed.
	shasum := sha256.New()
	_, _ = shasum.Write(pluginContents.Bytes())

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
	c.Ui.Say(fmt.Sprintf("Successfully installed plugin %s from %s to %s", args.PluginIdentifier, args.PluginPath, binaryPath))

	return 0
}
