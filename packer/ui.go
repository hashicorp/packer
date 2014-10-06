package packer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

type UiColor uint

const (
	UiColorRed     UiColor = 31
	UiColorGreen           = 32
	UiColorYellow          = 33
	UiColorBlue            = 34
	UiColorMagenta         = 35
	UiColorCyan            = 36
)

// The Ui interface handles all communication for Packer with the outside
// world. This sort of control allows us to strictly control how output
// is formatted and various levels of output.
type Ui interface {
	Ask(string) (string, error)
	Say(string)
	Message(string)
	Error(string)
	Machine(string, ...string)
}

// ColoredUi is a UI that is colored using terminal colors.
type ColoredUi struct {
	Color      UiColor
	ErrorColor UiColor
	Ui         Ui
}

// TargettedUi is a UI that wraps another UI implementation and modifies
// the output to indicate a specific target. Specifically, all Say output
// is prefixed with the target name. Message output is not prefixed but
// is offset by the length of the target so that output is lined up properly
// with Say output. Machine-readable output has the proper target set.
type TargettedUi struct {
	Target string
	Ui     Ui
}

// The BasicUI is a UI that reads and writes from a standard Go reader
// and writer. It is safe to be called from multiple goroutines. Machine
// readable output is simply logged for this UI.
type BasicUi struct {
	Reader      io.Reader
	Writer      io.Writer
	ErrorWriter io.Writer
	l           sync.Mutex
	interrupted bool
}

// MachineReadableUi is a UI that only outputs machine-readable output
// to the given Writer.
type MachineReadableUi struct {
	Writer io.Writer
}

func (u *ColoredUi) Ask(query string) (string, error) {
	return u.Ui.Ask(u.colorize(query, u.Color, true))
}

func (u *ColoredUi) Say(message string) {
	u.Ui.Say(u.colorize(message, u.Color, true))
}

func (u *ColoredUi) Message(message string) {
	u.Ui.Message(u.colorize(message, u.Color, false))
}

func (u *ColoredUi) Error(message string) {
	color := u.ErrorColor
	if color == 0 {
		color = UiColorRed
	}

	u.Ui.Error(u.colorize(message, color, true))
}

func (u *ColoredUi) Machine(t string, args ...string) {
	// Don't colorize machine-readable output
	u.Ui.Machine(t, args...)
}

func (u *ColoredUi) colorize(message string, color UiColor, bold bool) string {
	if !u.supportsColors() {
		return message
	}

	attr := 0
	if bold {
		attr = 1
	}

	return fmt.Sprintf("\033[%d;%dm%s\033[0m", attr, color, message)
}

func (u *ColoredUi) supportsColors() bool {
	// Never use colors if we have this environmental variable
	if os.Getenv("PACKER_NO_COLOR") != "" {
		return false
	}

	// For now, on non-Windows machine, just assume it does
	if runtime.GOOS != "windows" {
		return true
	}

	// On Windows, if we appear to be in Cygwin, then it does
	cygwin := os.Getenv("CYGWIN") != "" ||
		os.Getenv("OSTYPE") == "cygwin" ||
		os.Getenv("TERM") == "cygwin"

	return cygwin
}

func (u *TargettedUi) Ask(query string) (string, error) {
	return u.Ui.Ask(u.prefixLines(true, query))
}

func (u *TargettedUi) Say(message string) {
	u.Ui.Say(u.prefixLines(true, message))
}

func (u *TargettedUi) Message(message string) {
	u.Ui.Message(u.prefixLines(false, message))
}

func (u *TargettedUi) Error(message string) {
	u.Ui.Error(u.prefixLines(true, message))
}

func (u *TargettedUi) Machine(t string, args ...string) {
	// Prefix in the target, then pass through
	u.Ui.Machine(fmt.Sprintf("%s,%s", u.Target, t), args...)
}

func (u *TargettedUi) prefixLines(arrow bool, message string) string {
	arrowText := "==>"
	if !arrow {
		arrowText = strings.Repeat(" ", len(arrowText))
	}

	var result bytes.Buffer

	for _, line := range strings.Split(message, "\n") {
		result.WriteString(fmt.Sprintf("%s %s: %s\n", arrowText, u.Target, line))
	}

	return strings.TrimRightFunc(result.String(), unicode.IsSpace)
}

func (rw *BasicUi) Ask(query string) (string, error) {
	rw.l.Lock()
	defer rw.l.Unlock()

	if rw.interrupted {
		return "", errors.New("interrupted")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	log.Printf("ui: ask: %s", query)
	if query != "" {
		if _, err := fmt.Fprint(rw.Writer, query+" "); err != nil {
			return "", err
		}
	}

	result := make(chan string, 1)
	go func() {
		var line string
		if _, err := fmt.Fscanln(rw.Reader, &line); err != nil {
			log.Printf("ui: scan err: %s", err)
		}

		result <- line
	}()

	select {
	case line := <-result:
		return line, nil
	case <-sigCh:
		// Print a newline so that any further output starts properly
		// on a new line.
		fmt.Fprintln(rw.Writer)

		// Mark that we were interrupted so future Ask calls fail.
		rw.interrupted = true

		return "", errors.New("interrupted")
	}
}

func (rw *BasicUi) Say(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		log.Printf("[ERR] Failed to write to UI: %s", err)
	}
}

func (rw *BasicUi) Message(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		log.Printf("[ERR] Failed to write to UI: %s", err)
	}
}

func (rw *BasicUi) Error(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	writer := rw.ErrorWriter
	if writer == nil {
		writer = rw.Writer
	}

	log.Printf("ui error: %s", message)
	_, err := fmt.Fprint(writer, message+"\n")
	if err != nil {
		log.Printf("[ERR] Failed to write to UI: %s", err)
	}
}

func (rw *BasicUi) Machine(t string, args ...string) {
	log.Printf("machine readable: %s %#v", t, args)
}

func (u *MachineReadableUi) Ask(query string) (string, error) {
	return "", errors.New("machine-readable UI can't ask")
}

func (u *MachineReadableUi) Say(message string) {
	u.Machine("ui", "say", message)
}

func (u *MachineReadableUi) Message(message string) {
	u.Machine("ui", "message", message)
}

func (u *MachineReadableUi) Error(message string) {
	u.Machine("ui", "error", message)
}

func (u *MachineReadableUi) Machine(category string, args ...string) {
	now := time.Now().UTC()

	// Determine if we have a target, and set it
	target := ""
	commaIdx := strings.Index(category, ",")
	if commaIdx > -1 {
		target = category[0:commaIdx]
		category = category[commaIdx+1:]
	}

	// Prepare the args
	for i, v := range args {
		args[i] = strings.Replace(v, ",", "%!(PACKER_COMMA)", -1)
		args[i] = strings.Replace(args[i], "\r", "\\r", -1)
		args[i] = strings.Replace(args[i], "\n", "\\n", -1)
	}
	argsString := strings.Join(args, ",")

	_, err := fmt.Fprintf(u.Writer, "%d,%s,%s,%s\n", now.Unix(), target, category, argsString)
	if err != nil {
		if err == syscall.EPIPE || strings.Contains(err.Error(), "broken pipe") {
			// Ignore epipe errors because that just means that the file
			// is probably closed or going to /dev/null or something.
		} else {
			panic(err)
		}
	}
}
