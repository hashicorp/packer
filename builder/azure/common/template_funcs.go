package common

import (
	"bytes"
	"text/template"

	packertpl "github.com/hashicorp/packer/common/template"
)

func isValidByteValue(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	if 'a' <= b && b <= 'z' {
		return true
	}
	if 'A' <= b && b <= 'Z' {
		return true
	}
	return b == '.' || b == '_' || b == '-'
}

// Clean up image name by replacing invalid characters with "-"
// Names are not allowed to end in '.', '-', or  '_' and are trimmed.
func templateCleanImageName(s string) string {
	b := []byte(s)
	newb := make([]byte, len(b))
	for i := range newb {
		if isValidByteValue(b[i]) {
			newb[i] = b[i]
		} else {
			newb[i] = '-'
		}
	}

	newb = bytes.TrimRight(newb, "-_.")
	return string(newb)
}

var TemplateFuncs = template.FuncMap{
	"clean_resource_name": templateCleanImageName,
}
