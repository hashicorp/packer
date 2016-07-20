package reporters

import (
	"fmt"
	"path/filepath"

	"github.com/approvals/go-approval-tests/utils"
)

type quiet struct{}

// NewQuietReporter creates a new reporter that does nothing.
func NewQuietReporter() Reporter {
	return &quiet{}
}

func (s *quiet) Report(approved, received string) bool {
	approvedFull, _ := filepath.Abs(approved)
	receivedFull, _ := filepath.Abs(received)

	if utils.DoesFileExist(approved) {
		fmt.Printf("approval files did not match\napproved: %v\nreceived: %v\n", approvedFull, receivedFull)

	} else {
		fmt.Printf("result never approved\napproved: %v\nreceived: %v\n", approvedFull, receivedFull)
	}

	return true
}
