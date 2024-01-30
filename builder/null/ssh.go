// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package null

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		return host, nil
	}
}
