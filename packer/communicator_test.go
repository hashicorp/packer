package packer

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/iochan"
	"golang.org/x/sync/errgroup"
)

func TestRemoteCmd_StartWithUi(t *testing.T) {
	data := []string{
		"hello",
		"world",
		"foo",
		"there",
	}

	originalOutputReader, originalOutputWriter := io.Pipe()
	uilOutputReader, uilOutputWriter := io.Pipe()

	testComm := new(MockCommunicator)
	testComm.StartStdout = strings.Join(data, "\n") + "\n"
	testUi := &BasicUi{
		Reader: new(bytes.Buffer),
		Writer: uilOutputWriter,
	}

	rc := &RemoteCmd{
		Command: "test",
		Stdout:  originalOutputWriter,
	}
	ctx := context.TODO()

	wg := errgroup.Group{}

	testPrintFn := func(in io.Reader, expected []string) error {
		i := 0
		got := []string{}
		for output := range iochan.LineReader(in) {
			got = append(got, output)
			i++
			if i == len(expected) {
				// here ideally the LineReader chan should be closed, but since
				// the stream virtually has no ending we need to leave early.
				break
			}
		}
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Fatalf("bad output: %s", diff)
		}
		return nil
	}

	wg.Go(func() error { return testPrintFn(uilOutputReader, data) })
	wg.Go(func() error { return testPrintFn(originalOutputReader, data) })

	err := rc.RunWithUi(ctx, testComm, testUi)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	wg.Wait()
}

func TestRemoteCmd_Wait(t *testing.T) {
	var cmd RemoteCmd

	result := make(chan bool)
	go func() {
		cmd.Wait()
		result <- true
	}()

	cmd.SetExited(42)

	select {
	case <-result:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("never got exit notification")
	}
}
