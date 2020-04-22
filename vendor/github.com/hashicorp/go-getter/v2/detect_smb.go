package getter

import (
	"fmt"
	"strings"
)

// SmbDetector implements Detector to detect smb paths with //.
type SmbDetector struct{}

func (d *SmbDetector) Detect(src, pwd string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.HasPrefix(src, "//") {
		// This is a valid smb path and will be also checked as local file by the SmbGetter
		return fmt.Sprintf("smb:%s", src), true, nil
	}
	return "", false, nil
}
