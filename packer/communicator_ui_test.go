package packer

import (
	"cgl.tideland.biz/asserts"
	"fmt"
	"io"
	"testing"
	"time"
)

type StubCommunicator struct {
}

func (c *StubCommunicator) Start(r *RemoteCmd) error {
	go func() {
		io.WriteString(r.Stdout, r.Command)
	}()
	return nil
}

func (c *StubCommunicator) Upload(f string, r io.Reader) error {
	return nil
}

func (c *StubCommunicator) Download(f string, w io.Writer) error {
	return nil
}

type StubUi struct {
	Messaged bool
}

func (u *StubUi) Ask(query string) (string, error) {
	return "", nil
}

func (u *StubUi) Say(message string) {
}

func (u *StubUi) Message(message string) {
	u.Messaged = true
}

func (u *StubUi) Error(message string) {
}

func Test_StartWithUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cmd := RemoteCmd{Command: "hi"}
	var comm StubCommunicator
	var ui StubUi

	result := make(chan bool)
	go func() {
		if err := StartWithUi(&comm, &ui, &cmd); err != nil {
			err := fmt.Errorf("Unexpected error: %s", err)
			t.Fatal(err)
		}
		result <- true
	}()

	cmd.ExitStatus = 0
	cmd.Exited = true

	select {
	case <-result:
		assert.True(ui.Messaged, "should have written a message")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("never got exit notification")
	}
}
