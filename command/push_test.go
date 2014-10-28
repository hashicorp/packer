package command

import (
	"testing"
)

func TestPush_noArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run(nil)
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}

func TestPush_multiArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run([]string{"one", "two"})
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}
