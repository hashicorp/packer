// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func RegisterSecret(secret string) {
	if secret == "" {
		return
	}

	secrets := map[string]struct{}{
		secret: {},
	}

	normalized := strings.ReplaceAll(secret, "\r\n", "\n")
	secrets[normalized] = struct{}{}

	for _, line := range strings.Split(normalized, "\n") {
		if line == "" {
			continue
		}
		secrets[line] = struct{}{}
	}

	for value := range secrets {
		packersdk.LogSecretFilter.Set(value)
	}
}
