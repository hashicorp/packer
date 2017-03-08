package docker

import (
	"bufio"
	"io"
	"strings"

	"github.com/mitchellh/packer/packer"
)

func readAndStream(output io.Reader, ui packer.Ui) error {
	buf := bufio.NewReader(output)
	exitCh := make(chan error)

	go func() {
		var line string
		var err error

		for ; err == nil; line, err = buf.ReadString('\n') {
			if len(line) != 0 {
				ui.Message(strings.TrimSpace(line))
			}
		}
		exitCh <- err
	}()

	err := <-exitCh
	if err == io.EOF {
		return nil
	}
	return err
}
