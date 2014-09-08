package packer

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestConfigTemplateProcess_env(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = tpl.Process(`{{env "foo"}}`, nil)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestConfigTemplateProcess_isotime(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := tpl.Process(`{{isotime}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	val, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	currentTime := time.Now().UTC()
	if currentTime.Sub(val) > 2*time.Second {
		t.Fatalf("val: %d (current: %d)", val, currentTime)
	}
}

// Note must format with the magic Date: Mon Jan 2 15:04:05 -0700 MST 2006
func TestConfigTemplateProcess_isotime_withFormat(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Checking for a too-many arguments error
	// Because of the variadic function, compile time checking won't work
	_, err = tpl.Process(`{{isotime "20060102" "huh"}}`, nil)
	if err == nil {
		t.Fatalf("err: cannot have more than 1 input")
	}

	result, err := tpl.Process(`{{isotime "20060102"}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	ti := time.Now().UTC()
	val := fmt.Sprintf("%04d%02d%02d", ti.Year(), ti.Month(), ti.Day())

	if result != val {
		t.Fatalf("val: %s (formated: %s)", val, result)
	}
}

func TestConfigTemplateProcess_pwd(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := tpl.Process(`{{pwd}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != pwd {
		t.Fatalf("err: %s", result)
	}
}

func TestConfigTemplateProcess_timestamp(t *testing.T) {
	tpl, err := NewConfigTemplate()
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

	time.Sleep(2 * time.Second)

	result2, err := tpl.Process(`{{timestamp}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != result2 {
		t.Fatalf("bad: %#v %#v", result, result2)
	}
}

func TestConfigTemplateProcess_user(t *testing.T) {
	tpl, err := NewConfigTemplate()
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

func TestConfigTemplateProcess_uuid(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := tpl.Process(`{{uuid}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result) != 36 {
		t.Fatalf("err: %s", result)
	}
}

func TestConfigTemplateProcess_upper(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl.UserVars["foo"] = "bar"

	result, err := tpl.Process(`{{user "foo" | upper}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != "BAR" {
		t.Fatalf("bad: %s", result)
	}
}

func TestConfigTemplateProcess_lower(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl.UserVars["foo"] = "BAR"

	result, err := tpl.Process(`{{user "foo" | lower}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != "bar" {
		t.Fatalf("bad: %s", result)
	}
}

func TestConfigTemplateValidate(t *testing.T) {
	tpl, err := NewConfigTemplate()
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
