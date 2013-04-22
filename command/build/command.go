package build

import "fmt"
import "github.com/mitchellh/packer/packer"

type Command byte

func (Command) Run(env *packer.Environment, arg []string) int {
	fmt.Println("HI!")
	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from tempate"
}
