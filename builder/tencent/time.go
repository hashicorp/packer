package tencent

import (
	"fmt"
	"strconv"
	"time"
)

// CurrentTimeStamp returns the current time in Unix seconds as a string
func CurrentTimeStamp() string {
	ts := time.Now().Unix()
	s := strconv.FormatInt(ts, 10)
	return s
}

// SSHTimeStampSuffix returns the current time as yyyymmdd-hhmmss
func SSHTimeStampSuffix() string {
	now := time.Now()
	s := fmt.Sprintf("%d%02d%02d_%02d%02d", now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute())
	return s
}
