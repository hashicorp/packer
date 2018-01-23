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
			return &command.BuildCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"fix": func() (cli.Command, error) {
			return &command.FixCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"inspect": func() (cli.Command, error) {
			return &command.InspectCommand{
				Meta: *CommandMeta,
			}, nil
		},

		"push": func() (cli.Command, error) {
			return &command.PushCommand{
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

		"plugin": func() (cli.Command, error) {
			return &command.PluginCommand{
				Meta: *CommandMeta,
			}, nil
		},
	}
}
