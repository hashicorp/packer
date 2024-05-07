// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"

	getter "github.com/hashicorp/go-getter/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

var ErrInterrupted = errors.New("interrupted")

type UiColor uint

const (
	UiColorRed     UiColor = 31
	UiColorGreen           = 32
	UiColorYellow          = 33
	UiColorBlue            = 34
	UiColorMagenta         = 35
	UiColorCyan            = 36
)

// ColoredUi is a UI that is colored using terminal colors.
type ColoredUi struct {
	Color      UiColor
	ErrorColor UiColor
	Ui         packersdk.Ui
	PB         getter.ProgressTracker
}

var _ packersdk.Ui = new(ColoredUi)

func (u *ColoredUi) Ask(query string) (string, error) {
	return u.Ui.Ask(u.colorize(query, u.Color, true))
}

func (u *ColoredUi) Askf(query string, vals ...any) (string, error) {
	return u.Ask(fmt.Sprintf(query, vals...))
}

func (u *ColoredUi) Say(message string) {
	u.Ui.Say(u.colorize(message, u.Color, true))
}

func (u *ColoredUi) Sayf(message string, vals ...any) {
	u.Say(fmt.Sprintf(message, vals...))
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

func (u *ColoredUi) Errorf(message string, vals ...any) {
	u.Error(fmt.Sprintf(message, vals...))
}

func (u *ColoredUi) Machine(t string, args ...string) {
	// Don't colorize machine-readable output
	u.Ui.Machine(t, args...)
}

func (u *ColoredUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	return u.Ui.TrackProgress(u.colorize(src, u.Color, false), currentSize, totalSize, stream)
}

func (u *ColoredUi) colorize(message string, color UiColor, bold bool) string {
	if !u.supportsColors() {
		return message
	}

	attr := 0
	if bold {
		attr = 1
	}

	return fmt.Sprintf("\033[%d;%dm%s\033[0m", attr, color, message)
}

func (u *ColoredUi) supportsColors() bool {
	// Never use colors if we have this environmental variable
	if os.Getenv("PACKER_NO_COLOR") != "" {
		return false
	}

	// For now, on non-Windows machine, just assume it does
	if runtime.GOOS != "windows" {
		return true
	}

	// On Windows, if we appear to be in Cygwin, then it does
	cygwin := os.Getenv("CYGWIN") != "" ||
		os.Getenv("OSTYPE") == "cygwin" ||
		os.Getenv("TERM") == "cygwin"

	return cygwin
}

// TargetedUI is a UI that wraps another UI implementation and modifies
// the output to indicate a specific target. Specifically, all Say output
// is prefixed with the target name. Message output is not prefixed but
// is offset by the length of the target so that output is lined up properly
// with Say output. Machine-readable output has the proper target set.
type TargetedUI struct {
	Target string
	Ui     packersdk.Ui
}

var _ packersdk.Ui = new(TargetedUI)

func (u *TargetedUI) Ask(query string) (string, error) {
	return u.Ui.Ask(u.prefixLines(true, query))
}

func (u *TargetedUI) Askf(query string, args ...any) (string, error) {
	return u.Ask(fmt.Sprintf(query, args...))
}

func (u *TargetedUI) Say(message string) {
	u.Ui.Say(u.prefixLines(true, message))
}

func (u *TargetedUI) Sayf(message string, args ...any) {
	u.Say(fmt.Sprintf(message, args...))
}

func (u *TargetedUI) Message(message string) {
	u.Ui.Message(u.prefixLines(false, message))
}

func (u *TargetedUI) Error(message string) {
	u.Ui.Error(u.prefixLines(true, message))
}

func (u *TargetedUI) Errorf(message string, args ...any) {
	u.Error(fmt.Sprintf(message, args...))
}

func (u *TargetedUI) Machine(t string, args ...string) {
	// Prefix in the target, then pass through
	u.Ui.Machine(fmt.Sprintf("%s,%s", u.Target, t), args...)
}

func (u *TargetedUI) prefixLines(arrow bool, message string) string {
	arrowText := "==>"
	if !arrow {
		arrowText = strings.Repeat(" ", len(arrowText))
	}

	var result bytes.Buffer

	for _, line := range strings.Split(message, "\n") {
		result.WriteString(fmt.Sprintf("%s %s: %s\n", arrowText, u.Target, line))
	}

	return strings.TrimRightFunc(result.String(), unicode.IsSpace)
}

func (u *TargetedUI) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	return u.Ui.TrackProgress(u.prefixLines(false, src), currentSize, totalSize, stream)
}

// MachineReadableUi is a UI that only outputs machine-readable output
// to the given Writer.
type MachineReadableUi struct {
	Writer io.Writer
	PB     packersdk.NoopProgressTracker
}

var _ packersdk.Ui = new(MachineReadableUi)

func (u *MachineReadableUi) Ask(query string) (string, error) {
	return "", errors.New("machine-readable UI can't ask")
}

func (u *MachineReadableUi) Askf(query string, args ...any) (string, error) {
	return u.Ask(fmt.Sprintf(query, args...))
}

func (u *MachineReadableUi) Say(message string) {
	u.Machine("ui", "say", message)
}

func (u *MachineReadableUi) Sayf(message string, args ...any) {
	u.Say(fmt.Sprintf(message, args...))
}

func (u *MachineReadableUi) Message(message string) {
	u.Machine("ui", "message", message)
}

func (u *MachineReadableUi) Error(message string) {
	u.Machine("ui", "error", message)
}

func (u *MachineReadableUi) Errorf(message string, args ...any) {
	u.Error(fmt.Sprintf(message, args...))
}

func (u *MachineReadableUi) Machine(category string, args ...string) {
	now := time.Now().UTC()

	// Determine if we have a target, and set it
	target := ""
	commaIdx := strings.Index(category, ",")
	if commaIdx > -1 {
		target = category[0:commaIdx]
		category = category[commaIdx+1:]
	}

	// Prepare the args
	for i, v := range args {
		// Use packersdk.LogSecretFilter to scrub out sensitive variables
		args[i] = packersdk.LogSecretFilter.FilterString(args[i])
		args[i] = strings.Replace(v, ",", "%!(PACKER_COMMA)", -1)
		args[i] = strings.Replace(args[i], "\r", "\\r", -1)
		args[i] = strings.Replace(args[i], "\n", "\\n", -1)
	}
	argsString := strings.Join(args, ",")

	_, err := fmt.Fprintf(u.Writer, "%d,%s,%s,%s\n", now.Unix(), target, category, argsString)
	if err != nil {
		if err == syscall.EPIPE || strings.Contains(err.Error(), "broken pipe") {
			// Ignore epipe errors because that just means that the file
			// is probably closed or going to /dev/null or something.
		} else {
			panic(err)
		}
	}
	log.Printf("%d,%s,%s,%s\n", now.Unix(), target, category, argsString)
}

func (u *MachineReadableUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	return u.PB.TrackProgress(src, currentSize, totalSize, stream)
}

// TimestampedUi is a UI that wraps another UI implementation and
// prefixes each message with an RFC3339 timestamp
type TimestampedUi struct {
	Ui packersdk.Ui
	PB getter.ProgressTracker
}

var _ packersdk.Ui = new(TimestampedUi)

func (u *TimestampedUi) Ask(query string) (string, error) {
	return u.Ui.Ask(query)
}

func (u *TimestampedUi) Askf(query string, args ...any) (string, error) {
	return u.Ask(fmt.Sprintf(query, args...))
}

func (u *TimestampedUi) Say(message string) {
	u.Ui.Say(u.timestampLine(message))
}

func (u *TimestampedUi) Sayf(message string, args ...any) {
	u.Say(fmt.Sprintf(message, args...))
}

func (u *TimestampedUi) Message(message string) {
	u.Ui.Message(u.timestampLine(message))
}

func (u *TimestampedUi) Error(message string) {
	u.Ui.Error(u.timestampLine(message))
}

func (u *TimestampedUi) Errorf(message string, args ...any) {
	u.Error(fmt.Sprintf(message, args...))
}

func (u *TimestampedUi) Machine(message string, args ...string) {
	u.Ui.Machine(message, args...)
}

func (u *TimestampedUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	return u.Ui.TrackProgress(src, currentSize, totalSize, stream)
}

func (u *TimestampedUi) timestampLine(string string) string {
	return fmt.Sprintf("%v: %v", time.Now().Format(time.RFC3339), string)
}
