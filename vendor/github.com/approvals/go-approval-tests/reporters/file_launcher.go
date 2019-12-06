package reporters

import (
	"os/exec"
	"runtime"
)

type fileLauncher struct{}

// NewFileLauncherReporter launches registered application of the received file's type only.
func NewFileLauncherReporter() Reporter {
	return &fileLauncher{}
}

func (s *fileLauncher) Report(approved, received string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "Needed Title", received, "/B")
	default:
		cmd = exec.Command("open", received)
	}

	cmd.Start()
	return true
}
