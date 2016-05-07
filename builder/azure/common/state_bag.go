// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package common

import "github.com/mitchellh/multistep"

func IsStateCancelled(stateBag multistep.StateBag) bool {
	_, ok := stateBag.GetOk(multistep.StateCancelled)
	return ok
}
