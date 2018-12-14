package tencent

import (
	"strconv"
	"testing"
	"time"
)

func TestCurrentTimeStamp(t *testing.T) {
	ts1 := time.Now().Unix()
	s1 := strconv.FormatInt(ts1, 10)

	result := CurrentTimeStamp()

	ts2 := time.Now().Unix()
	s2 := strconv.FormatInt(ts2, 10)

	// the CurrentTimeStamp should be >= ts1 and <= ts2
	if !(result >= s1 && result <= s2) {
		t.Errorf("CurrentTimeStamp() = %s, expecting it to be >= %s and <= %s", result, s1, s2)
	}

}

func TestSSHTimeStampSuffix(t *testing.T) {
	s := SSHTimeStampSuffix()
	// 20180509-224303
	if len(s) != 15 {
		t.Fatalf("Length of SSHTimeStampSuffix isn't expected: len=%d, string=%s", len(s), s)
	}
}
