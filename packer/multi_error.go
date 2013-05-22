package packer

import (
	"fmt"
	"strings"
)

// MultiError is an error type to track multiple errors. This is used to
// accumulate errors in cases such as configuration parsing, and returning
// them as a single error.
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	points := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d error(s) occurred:\n\n%s",
		len(e.Errors), strings.Join(points, "\n"))
}
