// +build !windows

package gottyclient

import (
	"encoding/json"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"syscall"
)

func notifySignalSIGWINCH(c chan<- os.Signal) {
	signal.Notify(c, syscall.SIGWINCH)
}

func resetSignalSIGWINCH() {
	signal.Reset(syscall.SIGWINCH)
}

func syscallTIOCGWINSZ() ([]byte, error) {
	ws, err := unix.IoctlGetWinsize(0, 0)
	if err != nil {
		return nil, fmt.Errorf("ioctl error: %v", err)
	}
	b, err := json.Marshal(ws)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal error: %v", err)
	}
	return b, err
}
