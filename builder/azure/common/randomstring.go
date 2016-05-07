// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package common

import (
	"math/rand"
	"os"
	"time"
)

var pwSymbols = []string{
	"abcdefghijklmnopqrstuvwxyz",
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"0123456789",
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

func RandomString(chooseFrom string, length int) (randomString string) {
	cflen := len(chooseFrom)
	for i := 0; i < length; i++ {
		randomString += string(chooseFrom[rnd.Intn(cflen)])
	}
	return
}

func RandomPassword() (password string) {
	pwlen := 15
	batchsize := pwlen / len(pwSymbols)
	pw := make([]byte, 0, pwlen)
	// choose character set
	for c := 0; len(pw) < pwlen; c++ {
		s := RandomString(pwSymbols[c%len(pwSymbols)], rnd.Intn(batchsize-1)+1)
		pw = append(pw, []byte(s)...)
	}
	// truncate
	pw = pw[:pwlen]

	// permute
	for c := 0; c < pwlen-1; c++ {
		i := rnd.Intn(pwlen-c) + c
		x := pw[c]
		pw[c] = pw[i]
		pw[i] = x
	}
	return string(pw)
}
