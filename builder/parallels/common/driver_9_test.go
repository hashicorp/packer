package common

import (
	"testing"
)

func TestParallels9Driver_impl(t *testing.T) {
	var _ Driver = new(Parallels9Driver)
}
