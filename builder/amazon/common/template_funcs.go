package common

import (
	"bytes"
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
// For allowed characters see docs for Name parameter
// at http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateImage.html
func templateCleanAMIName(s string) string {
	allowed := []byte{'(', ')', '[', ']', ' ', '.', '/', '-', '\'', '@', '_'}
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

var TemplateFuncs = template.FuncMap{
	"clean_ami_name": templateCleanAMIName,
}
