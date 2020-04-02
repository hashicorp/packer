package random

import (
	"math/rand"
	"os"
	"time"
)

var (
	PossibleNumbers   = "0123456789"
	PossibleLowerCase = "abcdefghijklmnopqrstuvwxyz"
	PossibleUpperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	PossibleAlphaNum      = PossibleNumbers + PossibleLowerCase + PossibleUpperCase
	PossibleAlphaNumLower = PossibleNumbers + PossibleLowerCase
	PossibleAlphaNumUpper = PossibleNumbers + PossibleUpperCase
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

func Numbers(length int) string       { return String(PossibleNumbers, length) }
func AlphaNum(length int) string      { return String(PossibleAlphaNum, length) }
func AlphaNumLower(length int) string { return String(PossibleAlphaNumLower, length) }
func AlphaNumUpper(length int) string { return String(PossibleAlphaNumUpper, length) }

func String(chooseFrom string, length int) (randomString string) {
	cflen := len(chooseFrom)
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = chooseFrom[rnd.Intn(cflen)]
	}
	return string(bytes)
}
