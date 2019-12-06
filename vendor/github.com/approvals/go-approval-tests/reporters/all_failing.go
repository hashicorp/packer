package reporters

var (
	clipboardScratchData = ""
)

type allFailing struct{}

// NewAllFailingTestReporter copies move file command to your clipboard
func NewAllFailingTestReporter() Reporter {
	return &allFailing{}
}

func (s *allFailing) Report(approved, received string) bool {
	move := getMoveCommandText(approved, received)
	clipboardScratchData = clipboardScratchData + move + "\n"
	return copyToClipboard(clipboardScratchData)
}
