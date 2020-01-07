package command

import (
	"os"
	"strings"
)

func isDir(name string) (bool, error) {
	s, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func isHCLLoaded(name string) (bool, error) {
	if strings.HasSuffix(name, ".pkr.hcl") ||
		strings.HasSuffix(name, ".pkr.json") {
		return true, nil
	}
	return isDir(name)
}
