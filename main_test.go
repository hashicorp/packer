package main

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestExtractMachineReadable(t *testing.T) {
	var args, expected, result []string
	var mr bool

	// Not
	args = []string{"foo", "bar", "baz"}
	result, mr = extractMachineReadable(args)
	expected = []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}

	if mr {
		t.Fatal("should not be mr")
	}

	// Yes
	args = []string{"foo", "-machine-readable", "baz"}
	result, mr = extractMachineReadable(args)
	expected = []string{"foo", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}

	if !mr {
		t.Fatal("should be mr")
	}
}

func TestRandom(t *testing.T) {
	if rand.Intn(9999999) == 8498210 {
		t.Fatal("math.rand is not seeded properly")
	}
}
