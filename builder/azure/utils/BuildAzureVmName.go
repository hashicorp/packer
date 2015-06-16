// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

import (
	"math/rand"
	"os"
	"time"
	"sync"
)

const availableSymbols = "0123456789abcdefghijklmnopqrstuvwxyz"
const allowedVmNameLength = 15

var randmu sync.Mutex

func BuildAzureVmNameRandomSuffix(prefix string) (suffix string) {
	randmu.Lock()
	rand.Seed(time.Now().UnixNano() + int64(os.Getpid()))
	availableSymbolsLen := len(availableSymbols)
	genLen := allowedVmNameLength - len(prefix)
	for i := 0; i < genLen; i++ {
		suffix += string(availableSymbols[rand.Intn(availableSymbolsLen)])
	}
	randmu.Unlock()
	return
}
