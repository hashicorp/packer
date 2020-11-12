package iochan

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestLineReader(t *testing.T) {

	data := []string{"foo", "bar", "baz"}

	buf := new(bytes.Buffer)
	buf.WriteString(strings.Join(data, "\n") + "\n")

	ch := LineReader(buf)

	var result []string
	expected := data
	for v := range ch {
		result = append(result, v)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unexpected results: %#v", result)
	}
}
