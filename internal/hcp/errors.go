package hcp

import "fmt"

// BuildDone is the error retuned by an HCP handler when a build cannot be started since it's already marked as DONE.
type BuildDone struct {
	Message string
}

// Error returns the message for the BuildDone type
func (b BuildDone) Error() string {
	return fmt.Sprintf("BuildDone: %s", b.Message)
}
