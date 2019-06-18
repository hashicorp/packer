package utils

import (
	"errors"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

// Waiter to wait sth until it completed.
type Waiter interface {
	WaitForCompletion() error
	Cancel() error
}

// FuncWaiter used for waiting any condition function.
type FuncWaiter struct {
	Interval    time.Duration
	MaxAttempts int
	Checker     func() (bool, error)
	IgnoreError bool

	cancel chan struct{}
}

// WaitForCompletion will wait until the state of consdition is available.
// It will call the condition function to ensure state with interval.
func (w *FuncWaiter) WaitForCompletion() error {
	for i := 0; ; i++ {
		log.Infof("Waiting for completion ... attempted %v times, %v total", i, w.MaxAttempts)

		if i >= w.MaxAttempts {
			return errors.New("maximum attempts are reached")
		}

		if ok, err := w.Checker(); ok || (!w.IgnoreError && err != nil) {
			return err
		}

		select {
		case <-time.After(w.Interval):
			continue
		case <-w.cancel:
			break
		}
	}
}

// Cancel will stop all of WaitForCompletion function call.
func (w *FuncWaiter) Cancel() error {
	w.cancel <- struct{}{}
	return nil
}
