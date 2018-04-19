package bootcommand

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_chunkScanCodes(t *testing.T) {

	var chunktests = []struct {
		size int
		in   [][]string
		out  [][]string
	}{
		{
			3,
			[][]string{
				{"a", "b"},
				{"c"},
				{"d"},
				{"e", "f"},
				{"g", "h"},
				{"i", "j"},
				{"k"},
				{"l", "m"},
			},
			[][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
				{"g", "h"},
				{"i", "j", "k"},
				{"l", "m"},
			},
		},
		{
			-1,
			[][]string{
				{"a", "b"},
				{"c"},
				{"d"},
				{"e", "f"},
				{"g", "h"},
				{"i", "j"},
				{"k"},
				{"l", "m"},
			},
			[][]string{
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"},
			},
		},
	}

	for _, tt := range chunktests {
		out, err := chunkScanCodes(tt.in, tt.size)
		assert.NoError(t, err)
		assert.Equalf(t, tt.out, out, "expecting chunks of %d.", tt.size)
	}
}

func Test_chunkScanCodeError(t *testing.T) {
	// can't go from wider to thinner
	in := [][]string{
		{"a", "b", "c"},
		{"d", "e", "f"},
		{"g", "h"},
	}

	_, err := chunkScanCodes(in, 2)
	assert.Error(t, err)
}

func Test_pcxtSpecialLookup(t *testing.T) {
	in := "<rightShift><rightshiftoff><RIGHTSHIFTON>"
	expected := []string{"36", "b6", "b6", "36"}
	var codes []string
	sendCodes := func(c []string) error {
		codes = c
		return nil
	}
	d := NewPCXTDriver(sendCodes, -1)
	seq, err := GenerateExpressionSequence(in)
	assert.NoError(t, err)
	err = seq.Do(context.Background(), d)
	assert.NoError(t, err)
	assert.Equal(t, expected, codes)
}

func Test_flushes(t *testing.T) {
	in := "abc123<wait>098"
	expected := [][]string{
		{"1e", "9e", "30", "b0", "2e", "ae", "02", "82", "03", "83", "04", "84"},
		{"0b", "8b", "0a", "8a", "09", "89"},
	}
	var actual [][]string
	sendCodes := func(c []string) error {
		actual = append(actual, c)
		return nil
	}
	d := NewPCXTDriver(sendCodes, -1)
	seq, err := GenerateExpressionSequence(in)
	assert.NoError(t, err)
	err = seq.Do(context.Background(), d)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
