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
	Say(string)
	Error(string)
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

func (u *PrefixedUi) Say(message string) {
	u.Ui.Say(fmt.Sprintf("%s: %s", u.Prefix, message))
}

func (u *PrefixedUi) Error(message string) {
	u.Ui.Error(fmt.Sprintf("%s: %s", u.Prefix, message))
}

func (rw *ReaderWriterUi) Say(message string) {
	log.Printf("ui: %s", message)
	_, err := fmt.Fprint(rw.Writer, message+"\n")
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
