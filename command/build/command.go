package main

import "github.com/mitchellh/packer/packer"

type buildCommand byte

func (Command) Run(env packer.Environment, arg []string) int {
	env.Ui().Say("BUILDING!")
	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from tempate"
}
