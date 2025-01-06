# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

container {
	dependencies = false
	alpine_secdb = true
	secrets      = false
}

binary {
	secrets      = true
	go_modules   = true
	osv          = true
	oss_index    = true
	nvd          = false

	# Triage items that are _safe_ to ignore here. Note that this list should be
	# periodically cleaned up to remove items that are no longer found by the scanner.
	triage {
		suppress {
			vulnerabilities = [
				"GO-2022-0635", // github.com/aws/aws-sdk-go@v1.55.5 TODO(dduzgun-security): remove when deps is resolved
			]
		}
	}
}
