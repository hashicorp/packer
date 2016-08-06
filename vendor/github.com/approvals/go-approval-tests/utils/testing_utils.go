package utils

import "testing"

// AssertEqual Example:
// AssertEqual(t, 10, number, "number")
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf(message+"\n[expected != actual]\n[%s != %s]", expected, actual)
	}
}
