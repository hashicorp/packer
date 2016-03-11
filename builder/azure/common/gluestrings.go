// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package common

// removes overlap between the end of a and the start of b and
// glues them together
func GlueStrings(a, b string) string {
	shift := 0
	for shift < len(a) {
		i := 0
		for (i+shift < len(a)) && (i < len(b)) && (a[i+shift] == b[i]) {
			i++
		}
		if i+shift == len(a) {
			break
		}
		shift++
	}

	return string(a[:shift]) + b
}
