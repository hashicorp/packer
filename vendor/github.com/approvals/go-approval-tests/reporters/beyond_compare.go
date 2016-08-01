package reporters

type beyondCompare struct{}

// NewBeyondCompareReporter creates a new reporter for Beyond Compare 4.
func NewBeyondCompareReporter() Reporter {
	return &beyondCompare{}
}

func (s *beyondCompare) Report(approved, received string) bool {
	xs := []string{received, approved}
	programName := "C:/Program Files/Beyond Compare 4/BComp.exe"

	return launchProgram(programName, approved, xs...)
}
