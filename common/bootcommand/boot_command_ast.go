package bootcommand

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

/*
TODO:
	* pc-at abstraction
		* check that `<del>` works. It's different now.
		* parallels
		* hyperv-
*/

// KeysAction represents what we want to do with a key press.
// It can take 3 states. We either want to:
// * press the key once
// * press and hold
// * press and release
type KeyAction int

const (
	KeyOn KeyAction = 1 << iota
	KeyOff
	KeyPress
)

func (k KeyAction) String() string {
	switch k {
	case KeyOn:
		return "On"
	case KeyOff:
		return "Off"
	case KeyPress:
		return "Press"
	}
	panic(fmt.Sprintf("Unknwon KeyAction %d", k))
}

type expression interface {
	Do(context.Context, BCDriver) error
}

type expressionSequence []expression

// Do executes every expression in the sequence and then finalizes the driver.
func (s expressionSequence) Do(ctx context.Context, b BCDriver) error {
	for _, exp := range s {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := exp.Do(ctx, b); err != nil {
			return err
		}
	}
	return b.Finalize()
}

// GenerateExpressionSequence generates a sequence of expressions from the
// given command. This is the primary entry point to the boot command parser.
func GenerateExpressionSequence(command string) (expressionSequence, error) {
	got, err := ParseReader("", strings.NewReader(command))
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, fmt.Errorf("No expressions found.")
	}
	seq := expressionSequence{}
	for _, exp := range got.([]interface{}) {
		seq = append(seq, exp.(expression))
	}
	return seq, nil
}

type waitExpression struct {
	d time.Duration
}

// Do waits the amount of time described by the expression. It is cancellable
// through the context.
func (w *waitExpression) Do(ctx context.Context, _ BCDriver) error {
	log.Printf("[INFO] Waiting %s", w.d)
	select {
	case <-time.After(w.d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *waitExpression) String() string {
	return fmt.Sprintf("Wait<%s>", w.d)
}

type specialExpression struct {
	s      string
	action KeyAction
}

// Do sends the special command to the driver, along with the key action.
func (s *specialExpression) Do(ctx context.Context, driver BCDriver) error {
	return driver.SendSpecial(s.s, s.action)
}

func (s *specialExpression) String() string {
	return fmt.Sprintf("Spec-%s(%s)", s.action, s.s)
}

type literal struct {
	s      rune
	action KeyAction
}

// Do sends the key to the driver, along with the key action.
func (l *literal) Do(ctx context.Context, driver BCDriver) error {
	return driver.SendKey(l.s, l.action)
}

func (l *literal) String() string {
	return fmt.Sprintf("LIT-%s(%s)", l.action, string(l.s))
}
