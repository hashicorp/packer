package tencent

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"strings"
)

// Convert an interface{} to a []byte
func GetBytes(key interface{}) ([]byte, error) {
	//	gob.Register(map[string]interface{}{})
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	result := buf.Bytes()
	return result, nil
}

// IntToString converts an integer to a string
func IntToStr(value int) string {
	return strconv.Itoa(value)
}

func StrToInt(value string) int {
	return int(StrToInt64(value))
}

// Int64ToString converts an Int64 integer to a string
func Int64ToStr(value int64) string {
	return strconv.FormatInt(value, 10)
}

func StrToInt64(value string) int64 {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

// StrToBool converts a given string specifying "True", or "False" in various lower case or upper case
// into a boolean of true or false.
func StrToBool(value string) bool {
	switch strings.ToUpper(value) {
	case "TRUE":
		return true
	default:
		return false
	}
}

// BoolToStr converts a boolean into a string of either TRUE or FALSE.
func BoolToStr(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}
