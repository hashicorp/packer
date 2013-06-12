package packer

import (
	"fmt"
	"io"
	"log"
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
	Reader io.Reader
	Writer io.Writer
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

func (u *PrefixedUi) Say(message string) {
	u.Ui.Say(fmt.Sprintf("%s: %s", u.SayPrefix, message))
}

func (u *PrefixedUi) Message(message string) {
	u.Ui.Message(fmt.Sprintf("%s: %s", u.MessagePrefix, message))
}

func (u *PrefixedUi) Error(message string) {
	u.Ui.Error(fmt.Sprintf("%s: %s", u.SayPrefix, message))
}

func (rw *ReaderWriterUi) Say(message string) {
	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}

func (rw *ReaderWriterUi) Message(message string) {
	log.Printf("ui: %s", message)
	_, err := fmt.Fprintf(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}

func (rw *ReaderWriterUi) Error(message string) {
	log.Printf("ui error: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
	if err != nil {
		panic(err)
	}
}
