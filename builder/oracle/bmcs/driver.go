// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	client "github.com/mitchellh/packer/builder/oracle/bmcs/client"
)

// Driver interfaces between the builder steps and the BMCS SDK.
type Driver interface {
	CreateInstance(publicKey string) (string, error)
	CreateImage(id string) (client.Image, error)
	DeleteImage(id string) error
	GetInstanceIP(id string) (string, error)
	TerminateInstance(id string) error
	WaitForImageCreation(id string) error
	WaitForInstanceState(id string, waitStates []string, terminalState string) error
}
