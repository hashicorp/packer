package common

import (
	"testing"
)

func TestHTTPConfigPrepare_Bounds(t *testing.T) {
	// Test bad
	h := HTTPConfig{
		HTTPPortMin: 1000,
		HTTPPortMax: 500,
	}
	err := h.Prepare(nil)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test good
	h = HTTPConfig{
		HTTPPortMin: 0,
		HTTPPortMax: 0,
	}
	err = h.Prepare(nil)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	portMin := uint(8000)
	if h.HTTPPortMin != portMin {
		t.Fatalf("HTTPPortMin: expected %d got %d", portMin, h.HTTPPortMin)
	}
	portMax := uint(9000)
	if h.HTTPPortMax != portMax {
		t.Fatalf("HTTPPortMax: expected %d got %d", portMax, h.HTTPPortMax)
	}

	// Test good
	h = HTTPConfig{
		HTTPPortMin: 500,
		HTTPPortMax: 1000,
	}
	err = h.Prepare(nil)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
