package ncloud

import (
	"strconv"
)

func String(v string) *string {
	return &v
}

func IntString(n int) *string {
	return String(strconv.Itoa(n))
}

func StringInterfaceList(i []interface{}) []*string {
	vs := make([]*string, 0, len(i))
	for _, v := range i {
		switch v.(type) {
		case *string:
			vs = append(vs, v.(*string))
		default:
			vs = append(vs, String(v.(string)))
		}

	}
	return vs
}

func StringList(s []string) []*string {
	vs := make([]*string, 0, len(s))
	for _, v := range s {
		vs = append(vs, String(v))

	}
	return vs
}

func StringListValue(input []*string) []string {
	vs := make([]string, 0, len(input))
	for _, v := range input {
		vs = append(vs, StringValue(v))
	}
	return vs
}

func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

func Bool(v bool) *bool {
	return &v
}

func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

func Int(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v != nil {
		return *v
	}
	return 0
}

func Int32(v int32) *int32 {
	return &v
}

func Int32Value(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 0
}

func Int64(v int64) *int64 {
	return &v
}

func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

func Float32(v float32) *float32 {
	return &v
}

func Float32Value(v *float32) float32 {
	if v != nil {
		return *v
	}
	return 0
}
