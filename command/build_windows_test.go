package command

import "strings"

func init() {
	spaghettiCarbonara = fixWindowsLineEndings(spaghettiCarbonara)
	lasagna = fixWindowsLineEndings(lasagna)
}

func fixWindowsLineEndings(s string) string {
	return strings.ReplaceAll(s, "\n", " \r\n")
}
