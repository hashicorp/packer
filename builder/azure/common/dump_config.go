package common

import (
	"fmt"
	"github.com/mitchellh/reflectwalk"
	"reflect"
	"strings"
)

type walker struct {
	depth int
	say   func(string)
}

func newDumpConfig(say func(string)) *walker {
	return &walker{
		depth: 0,
		say:   say,
	}
}

func (s *walker) Enter(l reflectwalk.Location) error {
	s.depth += 1
	return nil
}

func (s *walker) Exit(l reflectwalk.Location) error {
	s.depth -= 1
	return nil
}

func (s *walker) Struct(v reflect.Value) error {
	return nil
}

func (s *walker) StructField(f reflect.StructField, v reflect.Value) error {
	if !s.shouldDump(v) {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		s.say(fmt.Sprintf("%s=%s", f.Name, s.formatValue(f.Name, v.String())))
	}

	return nil
}

func (s *walker) shouldDump(v reflect.Value) bool {
	return s.depth == 2 && v.IsValid() && v.CanInterface()
}

func (s *walker) formatValue(name, value string) string {
	if s.isMaskable(name) {
		return strings.Repeat("*", len(value))
	}

	return value
}

func (s *walker) isMaskable(name string) bool {
	up := strings.ToUpper(name)
	return strings.Contains(up, "SECRET") || strings.Contains(up, "PASSWORD")
}

func DumpConfig(config interface{}, say func(string)) {
	walker := newDumpConfig(say)
	reflectwalk.Walk(config, walker)
}
