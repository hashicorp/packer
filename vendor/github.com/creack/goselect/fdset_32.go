// +build darwin openbsd netbsd 386 arm mips mipsle

package goselect

// darwin, netbsd and openbsd uses uint32 on both amd64 and 386

const (
	// NFDBITS is the amount of bits per mask
	NFDBITS = 4 * 8
)
