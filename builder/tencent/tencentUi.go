package tencent

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/packer"
)

type myBasicPackerUiImpl struct {
}

func (ui *myBasicPackerUiImpl) Ask(Msg string) (string, error) {
	return "", nil
}
func (ui *myBasicPackerUiImpl) Say(Msg string) {
	fmt.Println(Msg)
}
func (ui *myBasicPackerUiImpl) Message(Msg string) {
	fmt.Print(Msg)
}
func (ui *myBasicPackerUiImpl) Error(Msg string) {
	os.Stderr.WriteString(Msg)
}
func (ui *myBasicPackerUiImpl) Machine(Msg string, Rest ...string) {
	fmt.Printf("%s %+v", Msg, Rest)
}

func NewPackerUi() packer.Ui {
	return &myBasicPackerUiImpl{}
}
