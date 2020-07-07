package command

import "strings"

func init() {
	spaghettiCarbonara = fixWindowsLineEndings(spaghettiCarbonara)
	lasagna = fixWindowsLineEndings(lasagna)
	tiramisu = fixWindowsLineEndings(tiramisu)
}

func fixWindowsLineEndings(s string) string {
	return strings.ReplaceAll(s, "\n", " \r\n")
}
