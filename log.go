// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bufio"
	"bytes"
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
	if logPath := os.Getenv(EnvLogFile); logPath != "" {
		var err error
		logOutput, err = os.Create(logPath)
		if err != nil {
			return nil, err
		}

		return logOutput, nil
	}

	// If we don't output logs to a file, or it wasn't requested, we can
	// return immediately without setting an output for logs.
	if os.Getenv(EnvLog) == "" || os.Getenv(EnvLog) == "0" {
		return
	}

	// no path; do a little light filtering to avoid double-dipping UI
	// calls.
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLinesSmallerThanBuffer)

	go func(scanner *bufio.Scanner) {
		defer w.Close()

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "ui:") {
				continue
			}
			if strings.Contains(scanner.Text(), "ui error:") {
				continue
			}
			fmt.Fprintf(os.Stderr, scanner.Text()+"\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "log output filter failed: %s\n", err)
		}
	}(scanner)
	logOutput = w

	return
}

// The below functions come from bufio.Scanner with a small tweak, to fix an
// edgecase where the default ScanFunc fails: sometimes, if someone tries to
// log a line that is longer than 64*1024 bytes long before it contains a
// newline, the ScanLine will continue to return, requesting more data from the
// buffer, which can't increase in size anymore, causing a hang.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanLinesSmallerThanBuffer(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}

	// Our tweak:
	// Buffer is full, so we can't get more data. Just return what we have as
	// its own token so we can keep going, even though there's no newline.
	if len(data)+1 >= bufio.MaxScanTokenSize {
		return len(data), data[0 : len(data)-1], nil
	}

	// Request more data.
	return 0, nil, nil
}
