// Copyright IBM Corp. 2024, 2025
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
	ErrorArgs   []any
	FatalCalled bool
	FatalArgs   []any
	SkipCalled  bool
	SkipArgs    []any

	f bool
}

func (t *mockT) Error(args ...any) {
	t.ErrorCalled = true
	t.ErrorArgs = args
	t.f = true
}

func (t *mockT) Fatal(args ...any) {
	t.FatalCalled = true
	t.FatalArgs = args
	t.f = true
}

func (t *mockT) Skip(args ...any) {
	t.SkipCalled = true
	t.SkipArgs = args
	t.f = true
}
