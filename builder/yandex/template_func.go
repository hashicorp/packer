package yandex

import (
	"strings"
	"text/template"
)

func isalphanumeric(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	if 'a' <= b && b <= 'z' {
		return true
	}
	return false
}

// Clean up resource name by replacing invalid characters with "-"
// and converting upper cases to lower cases
func templateCleanResourceName(s string) string {
	if reImageFamily.MatchString(s) {
		return s
	}
	b := []byte(strings.ToLower(s))
	newb := make([]byte, len(b))
	for i := range newb {
		if isalphanumeric(b[i]) {
			newb[i] = b[i]
		} else {
			newb[i] = '-'
		}
	}
	return string(newb)
}

var TemplateFuncs = template.FuncMap{
	"clean_resource_name": templateCleanResourceName,
}
