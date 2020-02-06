package getter

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

func SymlinkAny(oldname, newname string) error {
	sourcePath := filepath.FromSlash(oldname)

	// Use mklink to create a junction point
	output, err := exec.Command("cmd", "/c", "mklink", "/J", newname, sourcePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run mklink %v %v: %v %q", newname, sourcePath, err, output)
	}
	return nil
}

var ErrUnauthorized = syscall.ERROR_PRIVILEGE_NOT_HELD
