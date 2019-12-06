package reporters

import (
	"fmt"
)

type printSupportedDiffPrograms struct{}

// NewPrintSupportedDiffProgramsReporter creates a new reporter that states what reporters are supported.
func NewPrintSupportedDiffProgramsReporter() Reporter {
	return &quiet{}
}

func (s *printSupportedDiffPrograms) Report(approved, received string) bool {
	fmt.Printf("no diff reporters found on your system\ncurrently supported reporters are [in order of preference]:\nBeyond Compare\nIntelliJ")

	return false
}
