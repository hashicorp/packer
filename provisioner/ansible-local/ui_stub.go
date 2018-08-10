package ansiblelocal

import "github.com/hashicorp/packer/packer"

type uiStub struct{}

func (su *uiStub) Ask(string) (string, error) {
	return "", nil
}

func (su *uiStub) Error(string) {}

func (su *uiStub) Machine(string, ...string) {}

func (su *uiStub) Message(string) {}

func (su *uiStub) Say(msg string) {}
func (su *uiStub) GetProcessBar() packer.ProgressBar {
	return packer.GetDummyProgressBar()
}
