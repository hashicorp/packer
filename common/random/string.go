package random

import (
	"math/rand"
	"os"
	"time"
)

var (
	numbers   = "0123456789"
	lowerCase = "abcdefghijklmnopqrstuvwxyz"
	upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	alphaNum = numbers + lowerCase + upperCase
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

func AlphaNum(length int) string {
	return String(alphaNum, length)
}

func String(chooseFrom string, length int) (randomString string) {
	cflen := len(chooseFrom)
	for i := 0; i < length; i++ {
		randomString += string(chooseFrom[rnd.Intn(cflen)])
	}
	return
}
