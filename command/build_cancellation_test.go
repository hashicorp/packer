// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestBuildCommand_RunContext_CtxCancel(t *testing.T) {

	tests := []struct {
		name                 string
		args                 []string
		parallelPassingTests int
		expected             int
	}{
		{"cancel 1 pending build - parallel=true",
			[]string{"-parallel-builds=10", filepath.Join(testFixture("parallel"), "1lock-5wg.json")},
			5,
			1,
		},
		{"cancel in the middle with 2 pending builds - parallel=true",
			[]string{"-parallel-builds=10", filepath.Join(testFixture("parallel"), "2lock-4wg.json")},
			4,
			1,
		},
		{"cancel 1 locked build - debug - parallel=true",
			[]string{"-parallel-builds=10", "-debug=true", filepath.Join(testFixture("parallel"), "1lock.json")},
			0,
			1,
		},
		{"cancel 2 locked builds - debug - parallel=true",
			[]string{"-parallel-builds=10", "-debug=true", filepath.Join(testFixture("parallel"), "2lock.json")},
			0,
			1,
		},
		{"cancel 1 locked build - debug - parallel=false",
			[]string{"-parallel-builds=1", "-debug=true", filepath.Join(testFixture("parallel"), "1lock.json")},
			0,
			1,
		},
		{"cancel 2 locked builds - debug - parallel=false",
			[]string{"-parallel-builds=1", "-debug=true", filepath.Join(testFixture("parallel"), "2lock.json")},
			0,
			1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := NewParallelTestBuilder(tt.parallelPassingTests)
			locked := &LockedBuilder{unlock: make(chan interface{})}
			c := &BuildCommand{
				Meta: testMetaParallel(t, b, locked),
			}

			ctx, cancelCtx := context.WithCancel(context.Background())
			codeC := make(chan int)
			go func() {
				defer close(codeC)
				cfg, ret := c.ParseArgs(tt.args)
				if ret != 0 {
					t.Error("ParseArgs failed.")
					return
				}
				codeC <- c.RunContext(ctx, cfg)
			}()
			t.Logf("waiting for passing tests if any")
			b.wg.Wait() // ran `tt.parallelPassingTests` times
			t.Logf("cancelling context")
			cancelCtx()

			select {
			case code := <-codeC:
				if code != tt.expected {
					t.Logf("wrong code: %s", cmp.Diff(code, tt.expected))
					fatalCommand(t, c.Meta)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("deadlock")
			}
		})
	}
}
