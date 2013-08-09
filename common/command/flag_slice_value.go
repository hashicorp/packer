package command

import "strings"

// AppendSliceValue implements the flag.Value interface and allows multiple
// calls to the same variable to append a list.
type AppendSliceValue []string

func (s *AppendSliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *AppendSliceValue) Set(value string) error {
	if *s == nil {
		*s = make([]string, 0, 1)
	}

	*s = append(*s, value)
	return nil
}

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
