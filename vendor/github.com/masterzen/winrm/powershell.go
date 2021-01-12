package winrm

import (
	"encoding/base64"

	"golang.org/x/text/encoding/unicode"
)

// Powershell wraps a PowerShell script
// and prepares it for execution by the winrm client
func Powershell(psCmd string) string {
	// Disable unnecessary progress bars which considered as stderr.
	psCmd = "$ProgressPreference = 'SilentlyContinue';" + psCmd

	// Encode string to UTF16-LE
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	encoded, err := encoder.String(psCmd)
	if err != nil {
		return ""
	}

	// Finally make it base64 encoded which is required for powershell.
	psCmd = base64.StdEncoding.EncodeToString([]byte(encoded))

	// Specify powershell.exe to run encoded command
	return "powershell.exe -EncodedCommand " + psCmd
}
