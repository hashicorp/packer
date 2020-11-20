package shell

import "fmt"

func (p *Provisioner) ValidExitCode(code int) error {
	// Check exit code against allowed codes (likely just 0)
	validCodes := p.ValidExitCodes
	if len(validCodes) == 0 {
		validCodes = []int{0}
	}
	validExitCode := false
	for _, v := range validCodes {
		if code == v {
			validExitCode = true
			break
		}
	}
	if !validExitCode {
		return &ErrorInvalidExitCode{
			Code:    code,
			Allowed: validCodes,
		}
	}
	return nil
}

type ErrorInvalidExitCode struct {
	Code    int
	Allowed []int
}

func (e *ErrorInvalidExitCode) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Script exited with non-zero exit status: %d."+
		"Allowed exit codes are: %v",
		e.Code, e.Allowed)
}
