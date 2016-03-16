// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"time"
)

const (
	DeployCanceled  = "Canceled"
	DeployFailed    = "Failed"
	DeployDeleted   = "Deleted"
	DeploySucceeded = "Succeeded"
)

type DeploymentPoller struct {
	getProvisioningState func() (string, error)
	pause                func()
}

func NewDeploymentPoller(getProvisioningState func() (string, error)) *DeploymentPoller {
	pollDuration := time.Second * 15

	return &DeploymentPoller{
		getProvisioningState: getProvisioningState,
		pause:                func() { time.Sleep(pollDuration) },
	}
}

func (t *DeploymentPoller) PollAsNeeded() (string, error) {
	for {
		res, err := t.getProvisioningState()

		if err != nil {
			return res, err
		}

		switch res {
		case DeployCanceled, DeployDeleted, DeployFailed, DeploySucceeded:
			return res, nil
		default:
			break
		}

		t.pause()
	}
}
