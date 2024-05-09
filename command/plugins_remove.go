// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/hashicorp/packer/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/mitchellh/cli"
)

type PluginsRemoveCommand struct {
	Meta
}

func (c *PluginsRemoveCommand) Synopsis() string {
	return "Remove Packer plugins [matching a version]"
}

func (c *PluginsRemoveCommand) Help() string {
	helpText := `
Usage: packer plugins remove <plugin> [<version constraint>]

  This command will remove one or more installed Packer plugins.

  To remove a plugin matching a version constraint for the current OS and architecture.

      packer plugins remove github.com/hashicorp/happycloud v1.2.3

  To remove all versions of a plugin for the current OS and architecture omit the version constraint.

      packer plugins remove github.com/hashicorp/happycloud

  To remove a single plugin binary from the Packer plugin directory specify the absolute path to an installed binary. This syntax does not allow for version matching.

      packer plugins remove ~/.config/plugins/github.com/hashicorp/happycloud/packer-plugin-happycloud_v1.0.0_x5.0_linux_amd64
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsRemoveCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	return c.RunContext(ctx, args)
}

// deletePluginBinary removes a local plugin binary, and its related checksum file.
func deletePluginBinary(pluginPath string) error {
	if err := os.Remove(pluginPath); err != nil {
		return err
	}
	shasumFile := fmt.Sprintf("%s_SHA256SUM", pluginPath)

	if _, err := os.Stat(shasumFile); err != nil {
		log.Printf("[INFO] No SHA256SUM file to remove for the plugin, ignoring.")
		return nil
	}

	return os.Remove(shasumFile)
}

func (c *PluginsRemoveCommand) RunContext(buildCtx context.Context, args []string) int {
	if len(args) < 1 || len(args) > 2 {
		return cli.RunResultHelp
	}

	pluginDir, err := packer.PluginFolder()
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to get the plugin directory",
				Detail: fmt.Sprintf(
					"The directory in which plugins are installed could not be fetched from the environment. This is likely a Packer bug. Error: %s",
					err),
			},
		})
	}

	if filepath.IsAbs(args[0]) {
		if len(args) != 1 {
			c.Ui.Error("Unsupported: no version constraint may be specified with a local plugin path.\n")
			return cli.RunResultHelp
		}

		if !strings.Contains(args[0], pluginDir) {
			return writeDiags(c.Ui, nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid plugin location",
					Detail: fmt.Sprintf(
						"The path %q is not under the plugin directory inferred by Packer (%s) and will not be removed.",
						args[0],
						pluginDir),
				},
			})
		}

		log.Printf("will delete plugin located at %q", args[0])
		err := deletePluginBinary(args[0])
		if err != nil {
			return writeDiags(c.Ui, nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to delete plugin",
					Detail:   fmt.Sprintf("The plugin %q failed to be deleted with the following error: %q", args[0], err),
				},
			})
		}

		c.Ui.Say(args[0])

		return 0
	}

	opts := plugingetter.ListInstallationsOptions{
		PluginDirectory: c.Meta.CoreConfig.Components.PluginConfig.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:   runtime.GOOS,
			ARCH: runtime.GOARCH,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	plugin, err := addrs.ParsePluginSourceString(args[0])
	if err != nil {
		c.Ui.Errorf("Invalid source string %q: %s", args[0], err)
		return 1
	}

	// a plugin requirement that matches them all
	pluginRequirement := plugingetter.Requirement{
		Identifier: plugin,
	}

	if len(args) > 1 {
		constraints, err := version.NewConstraint(args[1])
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		pluginRequirement.VersionConstraints = constraints
	}

	installations, err := pluginRequirement.ListInstallations(opts)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	for _, installation := range installations {
		err := deletePluginBinary(installation.BinaryPath)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to remove plugin %q: %q", installation.BinaryPath, err))
			continue
		}
		c.Ui.Message(installation.BinaryPath)
	}

	if len(installations) == 0 {
		errMsg := fmt.Sprintf("No installed plugin found matching the plugin constraints %s", args[0])
		if len(args) == 2 {
			errMsg = fmt.Sprintf("%s %s", errMsg, args[1])
		}
		c.Ui.Error(errMsg)
		return 1
	}

	return 0
}
