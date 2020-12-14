// Package random is a helper for generating random alphanumeric strings.
package random

import (
	"math/rand"
	"os"
	"time"
)

var (
	PossibleNumbers          = "0123456789"
	PossibleLowerCase        = "abcdefghijklmnopqrstuvwxyz"
	PossibleUpperCase        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	PossibleSpecialCharacter = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	PossibleAlphaNum      = PossibleNumbers + PossibleLowerCase + PossibleUpperCase
	PossibleAlphaNumLower = PossibleNumbers + PossibleLowerCase
	PossibleAlphaNumUpper = PossibleNumbers + PossibleUpperCase
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

// Numbers returns a random numeric string of the given length
func Numbers(length int) string { return String(PossibleNumbers, length) }

// AlphaNum returns a random alphanumeric string of the given length. The
// returned string can contain both uppercase and lowercase letters.
func AlphaNum(length int) string { return String(PossibleAlphaNum, length) }

// AlphaNumLower returns a random alphanumeric string of the given length. The
// returned string can contain lowercase letters, but not uppercase.
func AlphaNumLower(length int) string { return String(PossibleAlphaNumLower, length) }

// AlphaNumUpper returns a random alphanumeric string of the given length. The
// returned string can contain uppercase letters, but not lowercase.
func AlphaNumUpper(length int) string { return String(PossibleAlphaNumUpper, length) }

// String returns a random string of the given length, using only the component
// characters provided in the "chooseFrom" string.
func String(chooseFrom string, length int) (randomString string) {
	cflen := len(chooseFrom)
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = chooseFrom[rnd.Intn(cflen)]
	}
	return string(bytes)
}
