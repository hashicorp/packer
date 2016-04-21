package main

import (
	"os"
	"os/signal"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/packer/command"
	"github.com/mitchellh/packer/version"
)

// Commands is the mapping of all the available Terraform commands.
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
				Meta:              *CommandMeta,
				Revision:          version.GitCommit,
				Version:           version.Version,
				VersionPrerelease: version.VersionPrerelease,
				CheckFunc:         commandVersionCheck,
			}, nil
		},

		"plugin": func() (cli.Command, error) {
			return &command.PluginCommand{
				Meta: *CommandMeta,
			}, nil
		},
	}
}

// makeShutdownCh creates an interrupt listener and returns a channel.
// A message will be sent on the channel for every interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
