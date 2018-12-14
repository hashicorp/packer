package tencent

import (
	"math"
	"reflect"
	"testing"
)

// Self contained test that tests IntToStr converts a number to a string successfully.
func TestIntToStr(t *testing.T) {
	type args struct {
		num int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Convert 1 to string", args{1}, "1"},
		{"Convert 2 to string", args{2}, "2"},
		{"Convert 2 to string", args{math.MinInt32}, "-2147483648"},
		{"Convert 2 to string", args{math.MaxInt32}, "2147483647"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntToStr(tt.args.num); got != tt.want {
				t.Errorf("IntToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBytes(t *testing.T) {
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBytes(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrToInt(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToInt(tt.args.value); got != tt.want {
				t.Errorf("StrToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Self contained test that ensures an int64 is converted to a string successfully
func TestInt64ToStr(t *testing.T) {
	type args struct {
		num int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// args below specifies the args type, not the args field
		{"Convert 1 to string", args{1}, "1"},
		{"Convert 2 to string", args{2}, "2"},
		{"Convert 2 to string", args{math.MinInt64}, "-9223372036854775808"},
		{"Convert 2 to string", args{math.MaxInt64}, "9223372036854775807"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64ToStr(tt.args.num); got != tt.want {
				t.Errorf("Int64ToString() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStrToInt64(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToInt64(tt.args.value); got != tt.want {
				t.Errorf("StrToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrToBool(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToBool(tt.args.value); got != tt.want {
				t.Errorf("StrToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolToStr(t *testing.T) {
	type args struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolToStr(tt.args.value); got != tt.want {
				t.Errorf("BoolToStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
