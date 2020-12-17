package packer

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

type secretFilter struct {
	s map[string]struct{}
	m sync.Mutex
	w io.Writer
}

func (l *secretFilter) Set(secrets ...string) {
	l.m.Lock()
	defer l.m.Unlock()
	for _, s := range secrets {
		l.s[s] = struct{}{}
	}
}

func (l *secretFilter) SetOutput(output io.Writer) {
	l.m.Lock()
	defer l.m.Unlock()
	l.w = output
}

func (l *secretFilter) Write(p []byte) (n int, err error) {
	for s := range l.s {
		if s != "" {
			p = bytes.Replace(p, []byte(s), []byte("<sensitive>"), -1)
		}
	}
	return l.w.Write(p)
}

// FilterString will overwrite any senstitive variables in a string, returning
// the filtered string.
func (l *secretFilter) FilterString(message string) string {
	for s := range l.s {
		if s != "" {
			message = strings.Replace(message, s, "<sensitive>", -1)
		}
	}
	return message
}

func (l *secretFilter) get() (s []string) {
	l.m.Lock()
	defer l.m.Unlock()
	for k := range l.s {
		s = append(s, k)
	}
	return
}

var LogSecretFilter secretFilter

func init() {
	LogSecretFilter.s = make(map[string]struct{})
}
