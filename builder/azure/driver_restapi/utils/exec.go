// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

import (
	"log"
	"os/exec"
	"fmt"
	"strings"
	"bytes"
)

func Exec(name string, arg ...string) error {

	log.Printf("Executing: %#v\n", arg)

	var stdout, stderr bytes.Buffer

	script := exec.Command(name, arg...)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	if _, ok := err.(*exec.ExitError); ok {
	err = fmt.Errorf("Exec error: %s\n", err)
	}

	stderrString := strings.TrimSpace(stderr.String())
	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("Exec stdout: %s\n", stdoutString)
	log.Printf("Exec stderr: %s\n", stderrString)
	if len(stderrString) > 0 {
	err = fmt.Errorf("%s\n", stderrString)
	}

	return err
}
