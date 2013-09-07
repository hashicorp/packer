package common

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"text/template"
)

func isalphanumeric(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	if 'a' <= b && b <= 'z' {
		return true
	}
	if 'A' <= b && b <= 'Z' {
		return true
	}
	return false
}

// Clean up AMI name by replacing invalid characters with "-"
func cleanAMIName(s string) string {
	allowed := []byte{'(', ')', ',', '/', '-', '_'}
	b := []byte(s)
	newb := make([]byte, len(b))
	for i, c := range b {
		if isalphanumeric(c) || bytes.IndexByte(allowed, c) != -1 {
			newb[i] = c
		} else {
			newb[i] = '-'
		}
	}
	return string(newb[:])
}

func templateCleanAMIName(args ...interface{}) string {
	s, ok := "", false
	if len(args) == 1 {
		s, ok = args[0].(string)
	}
	if !ok {
		s = fmt.Sprint(args...)
	}
	s = cleanAMIName(s)
	return s
}

func AddAMITemplateFuncs(t *packer.ConfigTemplate) {
	t.AddFuncs(template.FuncMap{
		"clean_ami_name": templateCleanAMIName,
	})
}
