package shell_local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToLinuxPath(t *testing.T) {
	winPath := "C:/path/to/your/file"
	winBashPath := "/mnt/c/path/to/your/file"
	converted, _ := ConvertToLinuxPath(winPath)
	assert.Equal(t, winBashPath, converted,
		"Should have converted %s to %s -- not %s", winPath, winBashPath, converted)

}
