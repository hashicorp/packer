package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// These are the environmental variables that determine if we log, and if
// we log whether or not the log should go to a file.
const EnvLog = "PACKER_LOG"          //Set to True
const EnvLogFile = "PACKER_LOG_PATH" //Set to a file

// logOutput determines where we should send logs (if anywhere).
func logOutput() (logOutput io.Writer, err error) {
	logOutput = nil
	if os.Getenv(EnvLog) != "" && os.Getenv(EnvLog) != "0" {
		logOutput = os.Stderr

		if logPath := os.Getenv(EnvLogFile); logPath != "" {
			var err error
			logOutput, err = os.Create(logPath)
			if err != nil {
				return nil, err
			}
		} else {
			// no path; do a little light filtering to avoid double-dipping UI
			// calls.
			r, w := io.Pipe()
			scanner := bufio.NewScanner(r)
			go func(scanner *bufio.Scanner) {
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "ui:") {
						continue
					}
					os.Stderr.WriteString(fmt.Sprintf(scanner.Text() + "\n"))
				}
			}(scanner)
			logOutput = w
		}
	}

	return
}
