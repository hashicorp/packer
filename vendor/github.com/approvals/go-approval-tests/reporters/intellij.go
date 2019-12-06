package reporters

type intellij struct{}

// NewIntelliJReporter creates a new reporter for IntelliJ.
func NewIntelliJReporter() Reporter {
	return &intellij{}
}

func (s *intellij) Report(approved, received string) bool {
	xs := []string{"diff", received, approved}
	programName := "C:/Program Files (x86)/JetBrains/IntelliJ IDEA 2016/bin/idea.exe"

	return launchProgram(programName, approved, xs...)
}
