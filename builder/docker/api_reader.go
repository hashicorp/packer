package docker

import (
	_ "github.com/mitchellh/iochan"
	"github.com/mitchellh/packer/packer"
	"io"
	_ "log"
)

func readAndStream(reader io.ReadCloser, ui packer.Ui) error {

	// TODO
	buf := make([]byte, 400)
	// type Msg struct {
	// 	Status string
	// }
	// var msg Msg

	// n := 0
	var err error
	for ; err != io.EOF; _, err = io.ReadAtLeast(reader, buf, 10) {
		// lines := bytes.Split(buf, []byte("\n"))

		// for l := range lines {
		// 	err = json.Unmarshal(lines[l], &msg)
		// 	if err != nil {
		// 		fmt.Printf("json err: %s\n", err.Error())
		// 	}
		// 	fmt.Printf("-> %s\n", msg.Status)
		// }
		// n, err = io.ReadAtLeast(reader, buf, 10)
		ui.Message(string(buf))
	}

	/*
		  stdout_r, stdout_w := io.Pipe()
			defer stdout_w.Close()

			cmd.Stdout = stdout_w

			// Create the channels we'll use for data
			stdoutCh := iochan.DelimReader(stdout_r, '\n')

			// Start the goroutine to watch for the exit
			go func() {
				defer stdout_w.Close()

				err := cmd.Wait()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitStatus = 1

					// There is no process-independent way to get the REAL
					// exit status so we just try to go deeper.
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						exitStatus = status.ExitStatus()
					}
				}

				exitCh <- exitStatus
			}()

			// This waitgroup waits for the streaming to end
			var streamWg sync.WaitGroup
			streamWg.Add(2)

			streamFunc := func(ch <-chan string) {
				defer streamWg.Done()

				for data := range ch {
					data = cleanOutputLine(data)
					if data != "" {
						ui.Message(data)
					}
				}
			}

			// Stream stderr/stdout
			go streamFunc(stdoutCh)

			// Wait for the process to end and then wait for the streaming to end
			exitStatus := <-exitCh
			streamWg.Wait()

			if exitStatus != 0 {
				return fmt.Errorf("Bad exit status: %d", exitStatus)
			}
	*/
	return nil
}
