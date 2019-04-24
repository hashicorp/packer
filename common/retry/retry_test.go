package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func success(context.Context) error { return nil }

func wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

var failErr = errors.New("woops !")

func fail(context.Context) error { return failErr }

type failOnce bool

func (ran *failOnce) Run(context.Context) error {
	if !*ran {
		*ran = true
		return failErr
	}
	return nil
}

func TestConfig_Run(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	type fields struct {
		StartTimeout time.Duration
		RetryDelay   func() time.Duration
		Tries        int
	}
	type args struct {
		ctx context.Context
		fn  func(context.Context) error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{"success",
			fields{StartTimeout: time.Second},
			args{context.Background(), success},
			nil},
		{"context cancelled",
			fields{StartTimeout: time.Second},
			args{cancelledCtx, wait},
			context.Canceled},
		{"timeout",
			fields{StartTimeout: 20 * time.Millisecond, RetryDelay: func() time.Duration { return 10 * time.Millisecond }},
			args{cancelledCtx, fail},
			failErr},
		{"success after one failure",
			fields{Tries: 2, RetryDelay: func() time.Duration { return 0 }},
			args{context.Background(), new(failOnce).Run},
			nil},
		{"fail after one failure",
			fields{Tries: 1, RetryDelay: func() time.Duration { return 0 }},
			args{context.Background(), new(failOnce).Run},
			&RetryExhaustedError{failErr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				StartTimeout: tt.fields.StartTimeout,
				RetryDelay:   tt.fields.RetryDelay,
				Tries:        tt.fields.Tries,
			}
			err := cfg.Run(tt.args.ctx, tt.args.fn)
			if diff := cmp.Diff(err, tt.wantErr, DeepAllowUnexported(RetryExhaustedError{}, errors.New(""))); diff != "" {
				t.Fatalf("Config.Run() unexpected error: %s", diff)
			}
		})
	}
}

func TestBackoff_Linear(t *testing.T) {
	b := Backoff{
		InitialBackoff: 2 * time.Minute,
		Multiplier:     2,
	}

	linear := (&b).Linear

	if linear() != 2*time.Minute {
		t.Fatal("first backoff should be 2 minutes")
	}

	if linear() != 4*time.Minute {
		t.Fatal("second backoff should be 4 minutes")
	}
}
