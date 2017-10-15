package googlecompute

import (
	"regexp"
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

// Clean up image name by replacing invalid characters with "-"
// truncate up to 63 length, convert to a lower case
func templateCleanAMIName(s string) string {
	re := regexp.MustCompile(`^[a-z][-a-z0-9]{0,61}[a-z0-9]$`)
	if re.MatchString(s) {
		return s
	}
	b := []byte(strings.ToLower(s))
	l := 63
	if len(b) < 63 {
		l = len(b)
	}
	newb := make([]byte, l)
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
	"clean_ami_name": templateCleanAMIName,
}
