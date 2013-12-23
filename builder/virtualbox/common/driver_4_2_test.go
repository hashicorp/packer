package common

import (
	"testing"
)

func TestVBox42Driver_impl(t *testing.T) {
	var _ Driver = new(VBox42Driver)
}
