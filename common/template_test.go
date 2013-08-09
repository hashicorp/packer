package common

import (
	"math"
	"strconv"
	"testing"
	"time"
)

func TestTemplateProcess_timestamp(t *testing.T) {
	tpl, err := NewTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := tpl.Process(`{{timestamp}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	val, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	currentTime := time.Now().UTC().Unix()
	if math.Abs(float64(currentTime-val)) > 10 {
		t.Fatalf("val: %d (current: %d)", val, currentTime)
	}
}

func TestTemplateProcess_user(t *testing.T) {
	tpl, err := NewTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl.UserVars["foo"] = "bar"

	result, err := tpl.Process(`{{user "foo"}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != "bar" {
		t.Fatalf("bad: %s", result)
	}
}

func TestTemplateValidate(t *testing.T) {
	tpl, err := NewTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Valid
	err = tpl.Validate(`{{user "foo"}}`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Invalid
	err = tpl.Validate(`{{idontexist}}`)
	if err == nil {
		t.Fatal("should have error")
	}
}
