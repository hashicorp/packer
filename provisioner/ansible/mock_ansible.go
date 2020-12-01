// +build !windows

package ansible

import (
	"fmt"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type provisionLogicTracker struct {
	setupAdapterCalled   bool
	executeAnsibleCalled bool
	happyPath            bool
}

func (l *provisionLogicTracker) setupAdapter(ui packersdk.Ui, comm packersdk.Communicator) (string, error) {
	l.setupAdapterCalled = true
	if l.happyPath {
		return "fakeKeyString", nil
	}
	return "", fmt.Errorf("chose sadpath")
}

func (l *provisionLogicTracker) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error {
	l.executeAnsibleCalled = true
	if l.happyPath {
		return fmt.Errorf("Chose sadpath")
	}
	return nil
}
