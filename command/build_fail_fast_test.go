// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func waitAtomicAtLeast(t *testing.T, v *int32, want int32) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		if atomic.LoadInt32(v) >= want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for count >= %d (have %d)", want, atomic.LoadInt32(v))
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestBuildFailFast(t *testing.T) {
	tests := []struct {
		name            string
		failFast        bool
		wantCompleted   int32
		wantAbortedMin  int32
		unlockRemaining bool
	}{
		{
			name:            "without fail-fast remaining builds complete",
			failFast:        false,
			wantCompleted:   2,
			wantAbortedMin:  0,
			unlockRemaining: true,
		},
		{
			name:            "with fail-fast remaining builds abort",
			failFast:        true,
			wantCompleted:   0,
			wantAbortedMin:  2,
			unlockRemaining: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			failing := &FailingBuilder{ready: make(chan struct{})}
			locked := &LockedBuilder{unlock: make(chan interface{})}
			c := &BuildCommand{
				Meta: testMetaParallelWithFail(t, NewParallelTestBuilder(0), locked, failing),
			}

			args := []string{
				"-parallel-builds=10",
				filepath.Join(testFixture("parallel"), "1fail-2lock.json"),
			}
			if tt.failFast {
				args = append([]string{"-fail-fast"}, args...)
			}

			ctx := context.Background()
			codeC := make(chan int, 1)
			go func() {
				cfg, ret := c.ParseArgs(args)
				if ret != 0 {
					t.Error("ParseArgs failed")
					codeC <- ret
					return
				}
				codeC <- c.RunContext(ctx, cfg)
			}()

			waitAtomicAtLeast(t, &locked.started, 2)
			close(failing.ready)

			if tt.unlockRemaining {
				close(locked.unlock)
			}

			select {
			case code := <-codeC:
				if code == 0 {
					fatalCommand(t, c.Meta)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("deadlock: remaining parallel builds were not aborted")
			}

			completed := atomic.LoadInt32(&locked.completed)
			aborted := atomic.LoadInt32(&locked.aborted)
			if completed != tt.wantCompleted {
				t.Fatalf("completed lock builds: got %d, want %d (aborted=%d)", completed, tt.wantCompleted, aborted)
			}
			if aborted < tt.wantAbortedMin {
				t.Fatalf("aborted lock builds: got %d, want at least %d (completed=%d)", aborted, tt.wantAbortedMin, completed)
			}
		})
	}
}
