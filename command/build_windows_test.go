// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import "strings"

func init() {
	spaghettiCarbonara = fixWindowsLineEndings(spaghettiCarbonara)
	lasagna = fixWindowsLineEndings(lasagna)
	tiramisu = fixWindowsLineEndings(tiramisu)
	one = fixWindowsLineEndings(one)
	two = fixWindowsLineEndings(two)
}

func fixWindowsLineEndings(s string) string {
	return strings.ReplaceAll(s, "\n", " \r\n")
}
