package main

import (
	"os"
	"os/signal"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/packer/command"
)

// Commands is the mapping of all the available Terraform commands.
var Commands map[string]cli.CommandFactory

// Ui is the cli.Ui used for communicating to the outside world.
var Ui cli.Ui

const ErrorPrefix = "e:"
const OutputPrefix = "o:"

func init() {
	Ui = &cli.PrefixedUi{
		AskPrefix:    OutputPrefix,
		OutputPrefix: OutputPrefix,
		InfoPrefix:   OutputPrefix,
		ErrorPrefix:  ErrorPrefix,
		Ui:           &cli.BasicUi{Writer: os.Stdout},
	}

	meta := command.Meta{
		EnvConfig: &EnvConfig,
		Ui:        Ui,
	}

	Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
				Meta: meta,
			}, nil
		},

		"fix": func() (cli.Command, error) {
			return &command.FixCommand{
				Meta: meta,
			}, nil
		},

		"inspect": func() (cli.Command, error) {
			return &command.InspectCommand{
				Meta: meta,
			}, nil
		},

		"push": func() (cli.Command, error) {
			return &command.PushCommand{
				Meta: meta,
			}, nil
		},

		"validate": func() (cli.Command, error) {
			return &command.ValidateCommand{
				Meta: meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:              meta,
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				CheckFunc:         commandVersionCheck,
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
