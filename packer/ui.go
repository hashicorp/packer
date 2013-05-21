package packer

import (
	"fmt"
	"io"
	"log"
)

// The Ui interface handles all communication for Packer with the outside
// world. This sort of control allows us to strictly control how output
// is formatted and various levels of output.
type Ui interface {
	Say(format string, a ...interface{})
	Error(format string, a ...interface{})
}

// PrefixedUi is a UI that wraps another UI implementation and adds a
// prefix to all the messages going out.
type PrefixedUi struct {
	Prefix string
	Ui     Ui
}

// The ReaderWriterUi is a UI that writes and reads from standard Go
// io.Reader and io.Writer.
type ReaderWriterUi struct {
	Reader io.Reader
	Writer io.Writer
}

func (u *PrefixedUi) Say(format string, a ...interface{}) {
	u.Ui.Say(fmt.Sprintf("%s: %s", u.Prefix, format), a...)
}

func (u *PrefixedUi) Error(format string, a ...interface{}) {
	u.Ui.Error(fmt.Sprintf("%s: %s", u.Prefix, format), a...)
}

func (rw *ReaderWriterUi) Say(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	log.Printf("ui: %s", output)
	_, err := fmt.Fprint(rw.Writer, output+"\n")
	if err != nil {
		panic(err)
	}
}

func (rw *ReaderWriterUi) Error(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	log.Printf("ui error: %s", output)
	_, err := fmt.Fprint(rw.Writer, output+"\n")
	if err != nil {
		panic(err)
	}
}
