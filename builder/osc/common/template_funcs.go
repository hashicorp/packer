package common

import "bytes"

func isalphanumeric(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	if 'a' <= b && b <= 'z' {
		return true
	}
	if 'A' <= b && b <= 'Z' {
		return true
	}
	return false
}

func templateCleanOMIName(s string) string {
	allowed := []byte{'(', ')', '[', ']', ' ', '.', '/', '-', '\'', '@', '_'}
	b := []byte(s)
	newb := make([]byte, len(b))
	for i, c := range b {
		if isalphanumeric(c) || bytes.IndexByte(allowed, c) != -1 {
			newb[i] = c
		} else {
			newb[i] = '-'
		}
	}
	return string(newb[:])
}
