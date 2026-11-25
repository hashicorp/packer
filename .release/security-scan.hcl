# Copyright IBM Corp. 2024, 2026
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
				"GO-2026-5932", // golang.org/x/crypto - only affects the unmaintained openpgp subpackage (no fixed version); not imported by Packer, confirmed absent from the binary via govulncheck/symbol scan.
				"GO-2026-6905", // github.com/blakesmith/ar - The package has been replaced with a fork sbomify/ar
			]
		}
	}
}
