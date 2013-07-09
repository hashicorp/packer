package packer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
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
}

// ColoredUi is a UI that is colored using terminal colors.
type ColoredUi struct {
	Color      UiColor
	ErrorColor UiColor
	Ui         Ui
}

// PrefixedUi is a UI that wraps another UI implementation and adds a
// prefix to all the messages going out.
type PrefixedUi struct {
	SayPrefix     string
	MessagePrefix string
	Ui            Ui
}

// The ReaderWriterUi is a UI that writes and reads from standard Go
// io.Reader and io.Writer.
type ReaderWriterUi struct {
	Reader      io.Reader
	Writer      io.Writer
	l           sync.Mutex
	interrupted bool
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

func (u *ColoredUi) colorize(message string, color UiColor, bold bool) string {
	attr := 0
	if bold {
		attr = 1
	}

	return fmt.Sprintf("\033[%d;%d;40m%s\033[0m", attr, color, message)
}

func (u *PrefixedUi) Ask(query string) (string, error) {
	return u.Ui.Ask(u.prefixLines(u.SayPrefix, query))
}

func (u *PrefixedUi) Say(message string) {
	u.Ui.Say(u.prefixLines(u.SayPrefix, message))
}

func (u *PrefixedUi) Message(message string) {
	u.Ui.Message(u.prefixLines(u.MessagePrefix, message))
}

func (u *PrefixedUi) Error(message string) {
	u.Ui.Error(u.prefixLines(u.SayPrefix, message))
}

func (u *PrefixedUi) prefixLines(prefix, message string) string {
	var result bytes.Buffer

	for _, line := range strings.Split(message, "\n") {
		result.WriteString(fmt.Sprintf("%s: %s\n", prefix, line))
	}

	return strings.TrimRightFunc(result.String(), unicode.IsSpace)
}

func (rw *ReaderWriterUi) Ask(query string) (string, error) {
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

func (rw *ReaderWriterUi) Say(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}

func (rw *ReaderWriterUi) Message(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}

func (rw *ReaderWriterUi) Error(message string) {
	rw.l.Lock()
	defer rw.l.Unlock()

	log.Printf("ui error: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}
