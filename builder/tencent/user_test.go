package tencent

import "testing"

func TestUserHomeDir(t *testing.T) {
	homedir := UserHomeDir()
	if !DirectoryExists(homedir) {
		t.Fatal("UserHomeDir() failed test")
	}
}
