// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/packer/command"
	"github.com/mitchellh/cli"
)

// Commands is the mapping of all the available Packer commands.
var Commands map[string]cli.CommandFactory

// CommandMeta is the Meta to use for the commands. This must be written
// before the CLI is started.
var CommandMeta *command.Meta

const ErrorPrefix = "e:"
const OutputPrefix = "o:"

func init() {
	Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{Meta: *CommandMeta}, nil
		},
		"console": func() (cli.Command, error) {
			return &command.ConsoleCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"fix": func() (cli.Command, error) {
			return &command.FixCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"fmt": func() (cli.Command, error) {
			return &command.FormatCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"hcl2_upgrade": func() (cli.Command, error) {
			return &command.HCL2UpgradeCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"init": func() (cli.Command, error) {
			return &command.InitCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"inspect": func() (cli.Command, error) {
			return &command.InspectCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugin": func() (cli.Command, error) {
			return &command.PluginCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugins": func() (cli.Command, error) {
			return &command.PluginsCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugins installed": func() (cli.Command, error) {
			return &command.PluginsInstalledCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugins install": func() (cli.Command, error) {
			return &command.PluginsInstallCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugins remove": func() (cli.Command, error) {
			return &command.PluginsRemoveCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"plugins required": func() (cli.Command, error) {
			return &command.PluginsRequiredCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"validate": func() (cli.Command, error) {
			return &command.ValidateCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:      *CommandMeta,
				CheckFunc: commandVersionCheck,
			}, nil
		},
	}
}
