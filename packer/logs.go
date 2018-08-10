package packer

import (
	"bytes"
	"io"
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
