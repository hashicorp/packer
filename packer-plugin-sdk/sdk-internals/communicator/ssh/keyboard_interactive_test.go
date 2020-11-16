package ssh

import (
	"io"
	"log"
	"reflect"
	"testing"
)

type MockTerminal struct {
	toSend       []byte
	bytesPerRead int
	received     []byte
}

func (c *MockTerminal) Read(data []byte) (n int, err error) {
	n = len(data)
	if n == 0 {
		return
	}
	if n > len(c.toSend) {
		n = len(c.toSend)
	}
	if n == 0 {
		return 0, io.EOF
	}
	if c.bytesPerRead > 0 && n > c.bytesPerRead {
		n = c.bytesPerRead
	}
	copy(data, c.toSend[:n])
	c.toSend = c.toSend[n:]
	return
}

func (c *MockTerminal) Write(data []byte) (n int, err error) {
	c.received = append(c.received, data...)
	return len(data), nil
}

func TestKeyboardInteractive(t *testing.T) {
	type args struct {
		user        string
		instruction string
		questions   []string
		echos       []bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "questions are none",
			args: args{
				questions: []string{},
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "input answer interactive",
			args: args{
				questions: []string{"this is question"},
			},
			want:    []string{"xxxx"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockTerminal{
				toSend:       []byte("xxxx\r\x1b[A\r"),
				bytesPerRead: 1,
			}
			f := KeyboardInteractive(c)
			got, err := f(tt.args.user, tt.args.instruction, tt.args.questions, tt.args.echos)

			if (err != nil) != tt.wantErr {
				t.Errorf("KeyboardInteractive error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyboardInteractive = %v, want %v", got, tt.want)
			}
			log.Printf("finish")
		})
	}
}
