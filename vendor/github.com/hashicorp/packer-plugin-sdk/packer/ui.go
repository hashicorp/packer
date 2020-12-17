package packer

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	getter "github.com/hashicorp/go-getter/v2"
)

type TTY interface {
	ReadString() (string, error)
	Close() error
}

// The Ui interface handles all communication for Packer with the outside
// world. This sort of control allows us to strictly control how output
// is formatted and various levels of output.
type Ui interface {
	Ask(string) (string, error)
	Say(string)
	Message(string)
	Error(string)
	Machine(string, ...string)
	// TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser)
	getter.ProgressTracker
}

var ErrInterrupted = errors.New("interrupted")

// BasicUI is an implementation of  Ui that reads and writes from a standard Go
// reader and writer. It is safe to be called from multiple goroutines. Machine
// readable output is simply logged for this UI.
type BasicUi struct {
	Reader      io.Reader
	Writer      io.Writer
	ErrorWriter io.Writer
	l           sync.Mutex
	interrupted bool
	TTY         TTY
	PB          getter.ProgressTracker
}

var _ Ui = new(BasicUi)

func (rw *BasicUi) Ask(query string) (string, error) {
	rw.l.Lock()
	defer rw.l.Unlock()

	if rw.interrupted {
		return "", ErrInterrupted
	}

	if rw.TTY == nil {
		return "", errors.New("no available tty")
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Printf("ui: ask: %s", query)
	if query != "" {
		if _, err := fmt.Fprint(rw.Writer, query+" "); err != nil {
			return "", err
		}
	}

	result := make(chan string, 1)
	go func() {
		line, err := rw.TTY.ReadString()
		if err != nil {
			log.Printf("ui: scan err: %s", err)
			return
		}
		result <- strings.TrimSpace(line)
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

		return "", ErrInterrupted
	}
}

func (rw *BasicUi) Say(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	// Use LogSecretFilter to scrub out sensitive variables
	message = LogSecretFilter.FilterString(message)

	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		log.Printf("[ERR] Failed to write to UI: %s", err)
	}
}

func (rw *BasicUi) Message(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	// Use LogSecretFilter to scrub out sensitive variables
	message = LogSecretFilter.FilterString(message)

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

	// Use LogSecretFilter to scrub out sensitive variables
	message = LogSecretFilter.FilterString(message)

	log.Printf("ui error: %s", message)
	_, err := fmt.Fprint(writer, message+"\n")
	if err != nil {
		log.Printf("[ERR] Failed to write to UI: %s", err)
	}
}

func (rw *BasicUi) Machine(t string, args ...string) {
	log.Printf("machine readable: %s %#v", t, args)
}

func (rw *BasicUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	return rw.PB.TrackProgress(src, currentSize, totalSize, stream)
}

// Safe is a UI that wraps another UI implementation and
// provides concurrency-safe access
type SafeUi struct {
	Sem chan int
	Ui  Ui
	PB  getter.ProgressTracker
}

var _ Ui = new(SafeUi)

func (u *SafeUi) Ask(s string) (string, error) {
	u.Sem <- 1
	ret, err := u.Ui.Ask(s)
	<-u.Sem

	return ret, err
}

func (u *SafeUi) Say(s string) {
	u.Sem <- 1
	u.Ui.Say(s)
	<-u.Sem
}

func (u *SafeUi) Message(s string) {
	u.Sem <- 1
	u.Ui.Message(s)
	<-u.Sem
}

func (u *SafeUi) Error(s string) {
	u.Sem <- 1
	u.Ui.Error(s)
	<-u.Sem
}

func (u *SafeUi) Machine(t string, args ...string) {
	u.Sem <- 1
	u.Ui.Machine(t, args...)
	<-u.Sem
}

func (u *SafeUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	u.Sem <- 1
	ret := u.Ui.TrackProgress(src, currentSize, totalSize, stream)
	<-u.Sem

	return ret
}

// NoopProgressTracker is a progress tracker
// that displays nothing.
type NoopProgressTracker struct{}

// TrackProgress returns stream
func (*NoopProgressTracker) TrackProgress(_ string, _, _ int64, stream io.ReadCloser) io.ReadCloser {
	return stream
}
