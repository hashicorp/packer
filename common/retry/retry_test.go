package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func success(context.Context) error { return nil }

func wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

var failErr = errors.New("woops !")

func fail(context.Context) error { return failErr }

func TestConfig_Run(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	type fields struct {
		StartTimeout time.Duration
		RetryDelay   func() time.Duration
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
			fields{StartTimeout: time.Second, RetryDelay: nil},
			args{context.Background(), success},
			nil},
		{"context cancelled",
			fields{StartTimeout: time.Second, RetryDelay: nil},
			args{cancelledCtx, wait},
			context.Canceled},
		{"timeout",
			fields{StartTimeout: 20 * time.Millisecond, RetryDelay: func() time.Duration { return 10 * time.Millisecond }},
			args{cancelledCtx, fail},
			failErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				StartTimeout: tt.fields.StartTimeout,
				RetryDelay:   tt.fields.RetryDelay,
			}
			if err := cfg.Run(tt.args.ctx, tt.args.fn); err != tt.wantErr {
				t.Fatalf("Config.Run() error = %v, wantErr %v", err, tt.wantErr)
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
