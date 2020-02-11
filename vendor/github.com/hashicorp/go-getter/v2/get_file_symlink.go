// +build !windows

package getter

import (
	"os"
)

var ErrUnauthorized = os.ErrPermission
var SymlinkAny = os.Symlink
