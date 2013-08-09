package command

import "strings"

// SliceValue implements the flag.Value interface and allows a list of
// strings to be given on the command line and properly parsed into a slice
// of strings internally.
type SliceValue []string

func (s *SliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *SliceValue) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}
