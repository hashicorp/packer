package bootcommand

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type event struct {
	u    uint32
	down bool
}

type sender struct {
	e []event
}

func (s *sender) KeyEvent(u uint32, down bool) error {
	if s != nil {
		s.e = append(s.e, event{u, down})
	}
	return nil
}

func Test_vncSpecialLookup(t *testing.T) {
	in := "<rightShift><rightshiftoff><RIGHTSHIFTON>"
	expected := []event{
		{0xFFE2, true},
		{0xFFE2, false},
		{0xFFE2, false},
		{0xFFE2, true},
	}
	s := &sender{}

	d := NewVNCDriver(noopReboot, s, time.Duration(0))
	seq, err := GenerateExpressionSequence(in)
	assert.NoError(t, err)
	err = seq.Do(context.Background(), d)
	assert.NoError(t, err)
	assert.Equal(t, expected, s.e)
}

func Test_vncIntervalNotGiven(t *testing.T) {
	s := &sender{}
	d := NewVNCDriver(noopReboot, s, time.Duration(0))
	assert.Equal(t, d.interval, time.Duration(100)*time.Millisecond)
}

func Test_vncIntervalGiven(t *testing.T) {
	s := &sender{}
	d := NewVNCDriver(noopReboot, s, time.Duration(5000)*time.Millisecond)
	assert.Equal(t, d.interval, time.Duration(5000)*time.Millisecond)
}

func Test_vncReboot(t *testing.T) {
	in := "abc123<wait>098<reboot>"
	rebootCalled := false
	var s *sender
	d := NewVNCDriver(func() error { rebootCalled = true; return nil }, s)
	seq, err := GenerateExpressionSequence(in)
	assert.NoError(t, err)
	err = seq.Do(context.Background(), d)
	assert.NoError(t, err)
	assert.True(t, rebootCalled, "reboot should have been called")
}
