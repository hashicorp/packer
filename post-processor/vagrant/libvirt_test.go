package vagrant

import (
	"fmt"
	"testing"
)

func assertSizeInMegabytes(t *testing.T, size string, expected uint64) {
	actual := sizeInMegabytes(size)
	if actual != expected {
		t.Fatalf("the size `%s` was converted to `%d` but expected `%d`", size, actual, expected)
	}
}

func Test_sizeInMegabytes_WithInvalidUnitMustPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a panic but got none")
		}
	}()

	sizeInMegabytes("1234x")
}

func Test_sizeInMegabytes_WithoutUnitMustDefaultToMegabytes(t *testing.T) {
	assertSizeInMegabytes(t, "1234", 1234)
}

func Test_sizeInMegabytes_WithBytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, fmt.Sprintf("%db", 1234*1024*1024), 1234)
	assertSizeInMegabytes(t, fmt.Sprintf("%dB", 1234*1024*1024), 1234)
	assertSizeInMegabytes(t, "1B", 0)
}

func Test_sizeInMegabytes_WithKiloBytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, fmt.Sprintf("%dk", 1234*1024), 1234)
	assertSizeInMegabytes(t, fmt.Sprintf("%dK", 1234*1024), 1234)
	assertSizeInMegabytes(t, "1K", 0)
}

func Test_sizeInMegabytes_WithMegabytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, "1234m", 1234)
	assertSizeInMegabytes(t, "1234M", 1234)
	assertSizeInMegabytes(t, "1M", 1)
}

func Test_sizeInMegabytes_WithGigabytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, "1234g", 1234*1024)
	assertSizeInMegabytes(t, "1234G", 1234*1024)
	assertSizeInMegabytes(t, "1G", 1*1024)
}

func Test_sizeInMegabytes_WithTerabytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, "1234t", 1234*1024*1024)
	assertSizeInMegabytes(t, "1234T", 1234*1024*1024)
	assertSizeInMegabytes(t, "1T", 1*1024*1024)
}

func Test_sizeInMegabytes_WithPetabytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, "1234p", 1234*1024*1024*1024)
	assertSizeInMegabytes(t, "1234P", 1234*1024*1024*1024)
	assertSizeInMegabytes(t, "1P", 1*1024*1024*1024)
}

func Test_sizeInMegabytes_WithExabytesUnit(t *testing.T) {
	assertSizeInMegabytes(t, "1234e", 1234*1024*1024*1024*1024)
	assertSizeInMegabytes(t, "1234E", 1234*1024*1024*1024*1024)
	assertSizeInMegabytes(t, "1E", 1*1024*1024*1024*1024)
}
