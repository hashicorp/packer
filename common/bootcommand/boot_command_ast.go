package bootcommand

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

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
	// Do executes the expression
	Do(context.Context, BCDriver) error
	// Validate validates the expression without executing it
	Validate() error
}

type expressionSequence []expression

// Do executes every expression in the sequence and then flushes remaining
// scancodes.
func (s expressionSequence) Do(ctx context.Context, b BCDriver) error {
	// validate should never fail here, since it should be called before
	// expressionSequence.Do. Only reason we don't panic is so we can clean up.
	if errs := s.Validate(); errs != nil {
		return fmt.Errorf("Found an invalid boot command. This is likely an error in Packer, so please open a ticket.")
	}

	for _, exp := range s {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := exp.Do(ctx, b); err != nil {
			return err
		}
	}
	return b.Flush()
}

// Validate tells us if every expression in the sequence is valid.
func (s expressionSequence) Validate() (errs []error) {
	for _, exp := range s {
		if err := exp.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return
}

// GenerateExpressionSequence generates a sequence of expressions from the
// given command. This is the primary entry point to the boot command parser.
func GenerateExpressionSequence(command string) (expressionSequence, error) {
	seq := expressionSequence{}
	if command == "" {
		return seq, nil
	}
	got, err := ParseReader("", strings.NewReader(command))
	if err != nil {
		return nil, err
	}
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
func (w *waitExpression) Do(ctx context.Context, driver BCDriver) error {
	driver.Flush()
	log.Printf("[INFO] Waiting %s", w.d)
	select {
	case <-time.After(w.d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Validate returns an error if the time is <= 0
func (w *waitExpression) Validate() error {
	if w.d <= 0 {
		return fmt.Errorf("Expecting a positive wait value. Got %s", w.d)
	}
	return nil
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

// Validate always passes
func (s *specialExpression) Validate() error {
	return nil
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

// Validate always passes
func (l *literal) Validate() error {
	return nil
}

func (l *literal) String() string {
	return fmt.Sprintf("LIT-%s(%s)", l.action, string(l.s))
}
