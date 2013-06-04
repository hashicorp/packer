package build

import "strings"

type stringSliceValue []string

func (s *stringSliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceValue) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}
