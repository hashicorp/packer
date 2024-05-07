// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package acctest

import (
	"os"
	"testing"
)

func init() {
	testTesting = true

	if err := os.Setenv(TestEnvVar, "1"); err != nil {
		panic(err)
	}
}

func TestTest_noEnv(t *testing.T) {
	// Unset the variable
	t.Setenv(TestEnvVar, "")

	mt := new(mockT)
	Test(mt, TestCase{})

	if !mt.SkipCalled {
		t.Fatal("skip not called")
	}
}

func TestTest_preCheck(t *testing.T) {
	called := false

	mt := new(mockT)
	Test(mt, TestCase{
		PreCheck: func() { called = true },
	})

	if !called {
		t.Fatal("precheck should be called")
	}
}

// mockT implements TestT for testing
type mockT struct {
	ErrorCalled bool
	ErrorArgs   []interface{}
	FatalCalled bool
	FatalArgs   []interface{}
	SkipCalled  bool
	SkipArgs    []interface{}

	f bool
}

func (t *mockT) Error(args ...interface{}) {
	t.ErrorCalled = true
	t.ErrorArgs = args
	t.f = true
}

func (t *mockT) Fatal(args ...interface{}) {
	t.FatalCalled = true
	t.FatalArgs = args
	t.f = true
}

func (t *mockT) Skip(args ...interface{}) {
	t.SkipCalled = true
	t.SkipArgs = args
	t.f = true
}
