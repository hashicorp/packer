// +build !windows

package ansible

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
)

type provisionLogicTracker struct {
	setupAdapterCalled   bool
	executeAnsibleCalled bool
	happyPath            bool
}

func (l *provisionLogicTracker) setupAdapter(ui packer.Ui, comm packer.Communicator) (string, error) {
	l.setupAdapterCalled = true
	if l.happyPath {
		return "fakeKeyString", nil
	}
	return "", fmt.Errorf("chose sadpath")
}

func (l *provisionLogicTracker) executeAnsible(ui packer.Ui, comm packer.Communicator, privKeyFile string) error {
	l.executeAnsibleCalled = true
	if l.happyPath {
		return fmt.Errorf("Chose sadpath")
	}
	return nil
}
