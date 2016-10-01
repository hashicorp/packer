package googlecompute

import (
	"strings"

	// To get test dependencies managed by Godeps
	_ "github.com/stretchr/testify/assert"
)

type Image struct {
	Licenses  []string
	Name      string
	ProjectId string
	SelfLink  string
	SizeGb    int64
}

func (i *Image) IsWindows() bool {
	for _, license := range i.Licenses {
		if strings.Contains(license, "windows") {
			return true
		}
	}
	return false
}
