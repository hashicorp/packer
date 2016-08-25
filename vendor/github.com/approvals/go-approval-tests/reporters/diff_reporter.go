package reporters

import (
	"os/exec"

	"github.com/approvals/go-approval-tests/utils"
)

// NewFrontLoadedReporter creates the default front loaded reporter.
func NewFrontLoadedReporter() *Reporter {
	tmp := NewFirstWorkingReporter(
		NewContinuousIntegrationReporter(),
	)

	return &tmp
}

// NewDiffReporter creates the default diff reporter.
func NewDiffReporter() *Reporter {
	tmp := NewFirstWorkingReporter(
		NewBeyondCompareReporter(),
		NewIntelliJReporter(),
		NewPrintSupportedDiffProgramsReporter(),
		NewQuietReporter(),
	)

	return &tmp
}

func launchProgram(programName, approved string, args ...string) bool {
	if !utils.DoesFileExist(programName) {
		return false
	}

	utils.EnsureExists(approved)

	cmd := exec.Command(programName, args...)
	cmd.Start()
	return true
}
