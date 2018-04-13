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
	* tests
	  * fix vbox tests
	* comments
	* lower-case specials
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

func onOffToAction(t string) KeyAction {
	if strings.EqualFold(t, "on") {
		return KeyOn
	} else if strings.EqualFold(t, "off") {
		return KeyOff
	}
	panic(fmt.Sprintf("Unknown state '%s'. Expecting On or Off.", t))
}

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

func (s expressionSequence) Do(ctx context.Context, b BCDriver) error {
	for _, exp := range s {
		if err := exp.Do(ctx, b); err != nil {
			return err
		}
	}
	return b.Finalize()
}

// GenerateExpressionSequence generates a sequence of expressions from the
// given command.
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

func (l *literal) Do(ctx context.Context, driver BCDriver) error {
	return driver.SendKey(l.s, l.action)
}

func (l *literal) String() string {
	return fmt.Sprintf("LIT-%s(%s)", l.action, string(l.s))
}
