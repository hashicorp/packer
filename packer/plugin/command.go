package plugin

import (
	"github.com/mitchellh/packer/packer"
	"log"
)

type cmdCommand struct {
	command packer.Command
	client  *Client
}

func (c *cmdCommand) Help() (result string) {
	defer func() {
		r := recover()
		c.checkExit(r, func() { result = "" })
	}()

	result = c.command.Help()
	return
}

func (c *cmdCommand) Run(e packer.Environment, args []string) (exitCode int) {
	defer func() {
		r := recover()
		c.checkExit(r, func() { exitCode = 1 })
	}()

	exitCode = c.command.Run(e, args)
	return
}

func (c *cmdCommand) Synopsis() (result string) {
	defer func() {
		r := recover()
		c.checkExit(r, func() {
			result = ""
		})
	}()

	result = c.command.Synopsis()
	return
}

func (c *cmdCommand) checkExit(p interface{}, cb func()) {
	if c.client.Exited() {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}
